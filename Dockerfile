FROM golang:1.19
WORKDIR /app
EXPOSE 8081
RUN git clone https://github.com/nitesh-repos/AccuKnox.git

RUN mv AccuKnox/rest_api/main.go main.go
RUN rm -rf AccuKnox
RUN go mod init rest_api
RUN go mod tidy
RUN go build main.go
CMD ./main

# CMD while true; do sleep 1; done
# docker build -t niteshnacc/accu_knox:v1 .;
# # docker tag notes_api_img niteshnacc/accu_knox:v1
# docker push niteshnacc/accu_knox:v1
# # docker commit notes_api_cont
# docker run -itd --name notes_api -p8081:80 niteshnacc/accu_knox:v1