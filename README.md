
# XM Golang Exercise v21.0.0

This is my solution for the exercise as stated in XM_Golang-Exercise_v21.0.0.pdf.

### Launch

To start application use shell command like `$ ./xm-exercise -pgurl postgresql://usr:pass@host/xm_db -opts 1,2`.

Command line arguments:
- pgurl: PostgreSQL connection URL
- opts: choose active options from the exercise (type 1 or 2 or 1,2); default 2

Note: Specifying through a command line argument is a bad decision for production because user credentials are logged in shell history.
However, it works for the demonstration purposes of this exercise.

### Database

Application uses PostgreSQL as a selected DB technology.

`Company` table was created using the following SQL statement:
    
    CREATE TABLE company (
        id bigserial primary key,
        name varchar,
        code varchar,
        country varchar,
        website varchar,
        phone varchar
    )

Note: All fields except `id` are nullable.

### Testing

Active DB connection is required for running tests. DB URL can be provided through the environment variable `PGX_TEST_DATABASE`.
Also, one could change DB URL in the test file `company/handlers_test.go`.

Disclaimer: For such a small application I decided to use a minimum of external libraries and relied mostly on standard library.
In such a scenario testing turned out to be the main "headache" (at least for me).

### Option 1

For some reason API at ipapi.co randomly responds with an error "Too Many Requests". Randomly means that sometimes request goes through,
sometimes it fails with the error. Such a behaviour might be connected to my location somehow.

Due to the aforementioned API error tests at `TestCyprusRequest` might fail.

### Optional Task

Not implemented