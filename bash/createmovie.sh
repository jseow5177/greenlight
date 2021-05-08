OUTPUTMESSAGE="***** Response *****"

# Should give json.SyntaxError
echo Sending XML as request body...
echo "$OUTPUTMESSAGE"
curl -i -d '<?xml version="1.0" encoding="UTF-8"?><note><to>Alex</to></note>' localhost:4000/v1/movies
echo

# Should give json.SyntaxError
echo Sending malformed JSON with trailing comma...
echo "$OUTPUTMESSAGE"
curl -i -d '{"title": "Moana", }' localhost:4000/v1/movies
echo

# Should give json.UnmarshalTypeError
echo Sending an JSON array instead of an object...
echo "$OUTPUTMESSAGE"
curl -i -d '["foo", "bar"]' localhost:4000/v1/movies
echo

# Should give json.UnmarshalTypeError
echo Sending a numeric \"title\" value instead of a string...
echo "$OUTPUTMESSAGE"
curl -i -d '{"title": 123}' localhost:4000/v1/movies
echo

# Should give io.EOF
echo Sending an empty request body...
echo "$OUTPUTMESSAGE"
curl -i -X POST localhost:4000/v1/movies
echo

# Should give "json: unknown field"
echo Sending JSON with unknown key \"rating\"
echo "$OUTPUTMESSAGE"
curl -i -d '{"title": "Moana", "rating": "PG"}' localhost:4000/v1/movies
echo

# Multiple JSON values instead of one
echo Sending two consecutive JSON data...
echo "$OUTPUTMESSAGE"
curl -i -d '{"title": "Moana"}{"title": "Top Gun"}' localhost:4000/v1/movies
echo

# Body contains garbage content after the first JSON value
echo Sending JSON with garbage value after...
echo "$OUTPUTMESSAGE"
curl -i -d '{"title": "Moana"} :~()' localhost:4000/v1/movies
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
