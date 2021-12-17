FROM node:12-alpine as build

WORKDIR /app
COPY package.json lib/index.js ./
COPY data ./data 
RUN npm install

FROM node:12-slim

COPY --from=build /app /
CMD ["index.js"]