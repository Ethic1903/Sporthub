# syntax=docker/dockerfile:1
FROM nginx:1.27-alpine

RUN apk add --no-cache gettext

WORKDIR /app

COPY nginx.conf /etc/nginx/nginx.conf
COPY config.js.tmpl /etc/nginx/templates/config.js.tmpl
COPY entrypoint.sh /entrypoint.sh
COPY index.html styles.css app.js /usr/share/nginx/html/

RUN chmod +x /entrypoint.sh

ENV FRONTEND_GATEWAY_URL=http://localhost:8080/graphql

EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
