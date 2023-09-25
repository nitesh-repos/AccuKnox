# To build the image follow below steps:
# Get Dockerfile from repo https://github.com/nitesh-repos/AccuKnox/
docker build -t niteshnacc/accu_knox:v1 .;
docker push niteshnacc/accu_knox:v1

# To run the container follow below steps:
docker run -itd --name notes_api -p8081:80 niteshnacc/accu_knox:v1
