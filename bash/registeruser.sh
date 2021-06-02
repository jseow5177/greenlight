INPUTMESSAGE="***** Request Body *****"
OUTPUTMESSAGE="***** Response *****"

# Register a new user with valid credentials
echo Registering a new user with valid credentials...
echo "$INPUTMESSAGE"
BODY='{"name": "Alice Smith", "email": "ALICE@example.com", "password": "pa55word"}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -w '\nTime: %{time_total}\n' -i -d "$BODY" localhost:4000/v1/users
echo

# Register a new user with valid credentials but duplicate email
echo Registering a new user with valid credentials but duplicate email...
echo "$INPUTMESSAGE"
BODY='{"name": "Alice Smith", "email": "alice@example.com", "password": "pa55word"}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -w '\nTime: %{time_total}\n' -i -d "$BODY" localhost:4000/v1/users
echo

# Register a new user with invalid credentials
echo Registering a new user with invalid credentials...
echo "$INPUTMESSAGE"
BODY='{"name": "", "email": "alice@invalid.", "password": "123"}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -w '\nTime: %{time_total}\n' -i -d "$BODY" localhost:4000/v1/users
echo
