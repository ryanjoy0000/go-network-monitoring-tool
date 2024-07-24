# build a new tiny docker image
FROM alpine

# create folder in root
RUN mkdir /app

# # copy files from current service folder to app folder 
COPY . /app

# copy executable from previous image to new image
COPY ./data-collector-service/dataCollectorApp /app

# provide defaults for an executing container
CMD ["/app/dataCollectorApp"]
