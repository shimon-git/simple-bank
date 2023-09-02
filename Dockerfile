FROM postgres:alpine

ENV POSTGRES_USER="shimon"

ENV POSTGRES_PASSWORD="ShimonTest123!"

VOLUME /Users/shimonyaniv/Desktop/golang/simple_bank/data:/var/lib/postgresql/data

EXPOSE 5432