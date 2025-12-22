const mysql = require('mysql2/promise');
const fs = require('fs');
const path = require('path');

async function initDatabase() {
  console.log('Connecting to database...');

  try {
    const connection = await mysql.createConnection({
      host: 'leasecalcs-development.c0xbv45gqu40.us-west-2.rds.amazonaws.com',
      port: 3306,
      user: 'woulder',
      password: 'j32JgmxzycbaoLet9F#9C%wFfN*RF98O',
      database: 'woulder',
      multipleStatements: true
    });

    console.log('Connected! Running schema...');

    const schemaPath = path.join(__dirname, '..', 'backend', 'internal', 'database', 'schema.sql');
    const schema = fs.readFileSync(schemaPath, 'utf8');

    await connection.query(schema);
    console.log('✓ Database schema initialized successfully!');

    // Verify tables were created
    const [tables] = await connection.query('SHOW TABLES');
    console.log('\nTables created:');
    tables.forEach(table => {
      console.log('  -', Object.values(table)[0]);
    });

    // Verify locations were inserted
    const [locations] = await connection.query('SELECT * FROM locations');
    console.log(`\n✓ ${locations.length} default locations inserted:`);
    locations.forEach(loc => {
      console.log(`  - ${loc.name} (${loc.latitude}, ${loc.longitude})`);
    });

    await connection.end();
    console.log('\nDatabase initialization complete!');
  } catch (error) {
    console.error('Error initializing database:', error.message);
    process.exit(1);
  }
}

initDatabase();
