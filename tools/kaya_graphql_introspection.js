#!/usr/bin/env node

/**
 * Kaya GraphQL API Introspection Tool
 * Analyzes the GraphQL schema and fetches sample data from Kaya climbing app
 */

const https = require('https');
const fs = require('fs');
const path = require('path');

const KAYA_GRAPHQL_URL = 'https://kaya-app.kayaclimb.com/graphql';

// GraphQL introspection query
const INTROSPECTION_QUERY = `
  query IntrospectionQuery {
    __schema {
      queryType { name }
      mutationType { name }
      types {
        name
        kind
        description
        fields {
          name
          description
          args {
            name
            type {
              name
              kind
              ofType {
                name
                kind
              }
            }
          }
          type {
            name
            kind
            ofType {
              name
              kind
            }
          }
        }
      }
    }
  }
`;

// Sample query for Leavenworth location
const LOCATION_QUERY = `
  query LocationQuery($slug: String!) {
    location(slug: $slug) {
      id
      name
      slug
      description
      latitude
      longitude
    }
  }
`;

function makeGraphQLRequest(query, variables = {}) {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({ query, variables });
    
    const options = {
      hostname: 'kaya-app.kayaclimb.com',
      port: 443,
      path: '/graphql',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
        'User-Agent': 'Woulder/0.1.0 (Research Tool)',
        'Accept': 'application/json'
      }
    };

    const req = https.request(options, (res) => {
      let data = '';

      res.on('data', (chunk) => {
        data += chunk;
      });

      res.on('end', () => {
        try {
          const parsed = JSON.parse(data);
          resolve({ statusCode: res.statusCode, headers: res.headers, data: parsed });
        } catch (e) {
          reject(new Error(`Failed to parse response: ${e.message}\nRaw data: ${data}`));
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    req.write(postData);
    req.end();
  });
}

async function main() {
  console.log('='.repeat(80));
  console.log('KAYA GRAPHQL API INTROSPECTION');
  console.log('='.repeat(80));
  console.log();

  const results = {
    introspection: null,
    locationTest: null,
    timestamp: new Date().toISOString(),
  };

  try {
    // 1. Test basic connectivity
    console.log('1. Testing basic GraphQL connectivity...');
    const basicTest = await makeGraphQLRequest('{__typename}');
    console.log(`   Status: ${basicTest.statusCode}`);
    if (basicTest.statusCode === 200) {
      console.log('   ✓ Successfully connected to GraphQL endpoint');
    } else {
      console.log(`   ✗ Unexpected status code: ${basicTest.statusCode}`);
    }
    console.log();

    // 2. Try introspection
    console.log('2. Attempting GraphQL schema introspection...');
    const schemaResult = await makeGraphQLRequest(INTROSPECTION_QUERY);
    
    if (schemaResult.data.errors) {
      console.error('   ✗ Introspection errors:', JSON.stringify(schemaResult.data.errors, null, 2));
      results.introspection = { error: schemaResult.data.errors };
    } else if (schemaResult.data.data) {
      const schema = schemaResult.data.data.__schema;
      console.log(`   ✓ Introspection successful!`);
      console.log(`   Query Type: ${schema.queryType?.name}`);
      console.log(`   Mutation Type: ${schema.mutationType?.name || 'None'}`);
      console.log(`   Total Types: ${schema.types.length}`);
      
      results.introspection = schema;

      // Find relevant types
      const relevantTypes = schema.types.filter(t => 
        !t.name.startsWith('__') && 
        ['OBJECT', 'INPUT_OBJECT'].includes(t.kind) &&
        (t.name.toLowerCase().includes('location') || 
         t.name.toLowerCase().includes('climb') ||
         t.name.toLowerCase().includes('route') ||
         t.name.toLowerCase().includes('ascent') ||
         t.name.toLowerCase().includes('tick') ||
         t.name.toLowerCase().includes('comment') ||
         t.name.toLowerCase().includes('destination'))
      );

      console.log();
      console.log('3. Relevant Types Found:');
      relevantTypes.forEach(type => {
        console.log(`\n   ${type.name} (${type.kind})`);
        if (type.description) {
          console.log(`   Description: ${type.description}`);
        }
        if (type.fields) {
          console.log('   Fields:');
          type.fields.slice(0, 10).forEach(field => {
            const typeName = field.type.name || field.type.ofType?.name || 'Unknown';
            console.log(`      - ${field.name}: ${typeName}`);
          });
          if (type.fields.length > 10) {
            console.log(`      ... and ${type.fields.length - 10} more fields`);
          }
        }
      });

      // Find Query type and its fields
      const queryType = schema.types.find(t => t.name === schema.queryType?.name);
      if (queryType && queryType.fields) {
        console.log('\n4. Available Query Operations:');
        queryType.fields.forEach(field => {
          console.log(`   - ${field.name}`);
          if (field.args && field.args.length > 0) {
            field.args.forEach(arg => {
              const typeName = arg.type.name || arg.type.ofType?.name || 'Unknown';
              console.log(`      arg: ${arg.name} (${typeName})`);
            });
          }
        });
      }
    }

    console.log('\n' + '='.repeat(80));
    console.log('5. Testing Sample Location Query (Leavenworth)...');
    console.log('='.repeat(80));
    
    const locationResult = await makeGraphQLRequest(LOCATION_QUERY, { 
      slug: 'Leavenworth-344933' 
    });
    
    if (locationResult.data.errors) {
      console.error('   ✗ Location query errors:', JSON.stringify(locationResult.data.errors, null, 2));
      results.locationTest = { error: locationResult.data.errors };
    } else if (locationResult.data.data) {
      console.log('   ✓ Location query successful!');
      console.log(JSON.stringify(locationResult.data.data, null, 2));
      results.locationTest = locationResult.data.data;
    }

    // Save results to file
    const outputDir = path.join(__dirname, '..', 'docs');
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }
    
    const outputFile = path.join(outputDir, 'kaya-api-analysis.json');
    fs.writeFileSync(outputFile, JSON.stringify(results, null, 2));
    
    console.log('\n' + '='.repeat(80));
    console.log('INTROSPECTION COMPLETE');
    console.log('='.repeat(80));
    console.log(`\nResults saved to: ${outputFile}`);
    console.log('\nNext steps:');
    console.log('1. Review the saved results file');
    console.log('2. Use browser DevTools to capture actual queries from the web app');
    console.log('3. Document query structures in docs/kaya-graphql-schema.md');

  } catch (error) {
    console.error('\n✗ Error:', error.message);
    if (error.code) {
      console.error('  Code:', error.code);
    }
    process.exit(1);
  }
}

main();
