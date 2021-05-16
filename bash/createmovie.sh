INPUTMESSAGE="***** Request Body *****"
OUTPUTMESSAGE="***** Response *****"

# Should give json.SyntaxError
echo Sending XML as request body...
echo "$INPUTMESSAGE"
BODY='<?xml version="1.0" encoding="UTF-8"?><note><to>Alex</to></note>'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Should give json.SyntaxError
echo Sending malformed JSON with trailing comma...
echo "$INPUTMESSAGE"
BODY='{"title": "Moana", }'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Should give json.UnmarshalTypeError
echo Sending an JSON array instead of an object...
echo "$INPUTMESSAGE"
BODY='["foo", "bar"]'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Should give json.UnmarshalTypeError
echo Sending a numeric \"title\" value instead of a string...
echo "$INPUTMESSAGE"
BODY='{"title": 123}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Should give io.EOF
echo Sending an empty request body...
echo "$OUTPUTMESSAGE"
curl -i -X POST localhost:4000/v1/movies
echo

# Should give "json: unknown field"
echo Sending JSON with unknown key \"rating\"...
echo "$INPUTMESSAGE"
BODY='{"title": "Moana", "rating": "PG"}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Multiple JSON values instead of one
echo Sending two consecutive JSON data...
echo "$INPUTMESSAGE"
BODY='{"title": "Moana"}{"title": "Top Gun"}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Body contains garbage content after the first JSON value
echo Sending JSON with garbage value after...
echo "$INPUTMESSAGE"
BODY='{"title": "Moana"} :~()'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Send a 1.5MB JSON file
echo Sending a 1.5MB JSON file...
## Download dummy large file if does not exist
FILE=/tmp/largefile.json
if [ ! -f "$FILE" ]; then
  wget -O /tmp/largefile.json https://www.alexedwards.net/static/largefile.json
fi
echo "$OUTPUTMESSAGE"
curl -i -d @/tmp/largefile.json localhost:4000/v1/movies
echo

# Test validation logic of handler
echo Sending a JSON with invalid data...
echo "$INPUTMESSAGE"
BODY='{"title":"","year":1000,"runtime":"-123 mins","genres":["sci-fi","sci-fi"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Valid request
echo Sending a JSON with valid data...
echo "$INPUTMESSAGE"
BODY='{"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","adventure"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -i -d "$BODY" localhost:4000/v1/movies
echo

# Create more movies
echo Creating Deadpool movie...
echo "$INPUTMESSAGE"
BODY='{"title":"Deadpool","year":2016, "runtime":"108 mins","genres":["action","comedy"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -d "$BODY" localhost:4000/v1/movies

echo Creating Black Panther movie...
echo "$INPUTMESSAGE"
BODY='{"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["action","adventure"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -d "$BODY" localhost:4000/v1/movies

echo Creating The Breakfast Club movie...
echo "$INPUTMESSAGE"
BODY='{"title":"The Breakfast Club","year":1986, "runtime":"96 mins","genres":["drama"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -d "$BODY" localhost:4000/v1/movies

