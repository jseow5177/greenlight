INPUTMESSAGE="***** Request Body *****"
OUTPUTMESSAGE="***** Response *****"

# Test update by providing full body with valid data
echo Updating movie Black Panther...
echo "$INPUTMESSAGE"
BODY='{"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["action","adventure","sci-fi"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -X PATCH -d "$BODY" localhost:4000/v1/movies/3

# Test update by providing partial body with valid data
echo Updating movie The Breakfast Club...
echo "$INPUTMESSAGE"
BODY='{"year": 1985}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -X PATCH -d "$BODY" localhost:4000/v1/movies/4

# Test invalid update by providing partial body with invalid data
echo Updating movie The Breakfast Club...
echo "$INPUTMESSAGE"
BODY='{"year": 1985,"title": ""}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -X PATCH -d "$BODY" localhost:4000/v1/movies/4

# Special case: Client explicitly supplies a field in the JSON request with the value null
# null value in JSON is treated specially by Go. It will be unmarshaled into nil
# In this case, the handler will ignore the field and treat it like it has not been supplied
# However, the version number will still be incremented
echo Updating movie The Breakfast Club...
echo "$INPUTMESSAGE"
BODY='{"year": null,"title": null"}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -X PATCH -d "$BODY" localhost:4000/v1/movies/4