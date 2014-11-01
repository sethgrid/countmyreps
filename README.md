countmyreps
===========
To run tests, enter root directory of countmyreps, and run 'phpunit tests'. Requires PHPUnit > 3.7 and PHPUnit/DbUnit.
Because we are testing against a data model, we will need DbUnit and a test database. Import the sql in the setup directory into
a database for production and a test database called 'testdb'. Currently, DataStoreTest.php is the only test that requires DbUnit and
the testdb database.


Dependencies for Keen
=====================
Use composer to install keen
```bash
php bin/composer.phar install
```

Env
===
sample.conf contains all env vars that should be set to make app work