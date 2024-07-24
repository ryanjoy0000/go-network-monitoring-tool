# build a new tiny docker image
FROM alpine

# create folder in root
RUN mkdir /app

# copy executable from previous image to new image
COPY centralApp /app

# provide defaults for an executing container
CMD ["/app/centralApp"]
