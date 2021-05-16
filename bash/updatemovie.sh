INPUTMESSAGE="***** Request Body *****"
OUTPUTMESSAGE="***** Response *****"

# Update a movie of id 3
echo Updating movie Black Panther...
echo "$INPUTMESSAGE"
BODY='{"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["action","adventure","sci-fi"]}'
echo "$BODY"
echo "$OUTPUTMESSAGE"
curl -X PUT -d "$BODY" localhost:4000/v1/movies/3