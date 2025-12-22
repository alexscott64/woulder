require('dotenv').config({ path: require('path').join(__dirname, '..', 'backend', '.env') });
const mysql = require('mysql2/promise');
const fs = require('fs');
const path = require('path');

async function initDatabase() {
  console.log('Connecting to database...');

  // Validate required environment variables
  const required = ['DB_HOST', 'DB_PORT', 'DB_USER', 'DB_PASSWORD', 'DB_NAME'];
  const missing = required.filter(key => !process.env[key]);

  if (missing.length > 0) {
    console.error(`Error: Missing required environment variables: ${missing.join(', ')}`);
    console.error('Make sure backend/.env file exists and contains all required variables.');
    process.exit(1);
  }

  try {
    const connection = await mysql.createConnection({
      host: process.env.DB_HOST,
      port: parseInt(process.env.DB_PORT),
      user: process.env.DB_USER,
      password: process.env.DB_PASSWORD,
      database: process.env.DB_NAME,
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
