FROM node:24 AS builder

WORKDIR /app
# Set the build argument and environment variable
ARG PROD
ENV PROD=$PROD

COPY ./frontend .

COPY ./backend/api/api.yaml api.yaml
RUN sed -i "s|http://localhost:8080/api|$PROD/api|g" api.yaml

RUN npm install

RUN ./node_modules/.bin/openapi --input ./api.yaml --output ./src/lib/api --request ./src/lib/request.tmp

RUN npm run build 

FROM nginx:alpine AS runner

COPY --from=builder /app/out /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

