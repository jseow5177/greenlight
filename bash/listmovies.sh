INPUTMESSAGE="***** Query Params *****"
OUTPUTMESSAGE="***** Response *****"

echo Listing movies without query string parameters...
echo "$OUTPUTMESSAGE"
curl -i localhost:4000/v1/movies
echo

echo Invalid query string parameters...
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?page=abc&page_size=-1&sort=abc,-def"
echo

echo Listing movies where the title is \'black panther\'...
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?title=black+panther"
echo

echo Listing movies where the genres includes \'adventure\'
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?genres=adventure"
echo

echo Listing movies where the genres include both \'animation\' AND \'adventure\'
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?genres=adventure,animation"
echo

echo Listing movies where the title is \'iron man\'...
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?title=iron+man"
echo

echo Listing movies where the title contains the word \'panther\'...
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?title=panther"
echo

echo Listing movies where the title contains the words \'the club\'...
echo "$OUTPUTMESSAGE"
curl -i "localhost:4000/v1/movies?title=the+club"
echo