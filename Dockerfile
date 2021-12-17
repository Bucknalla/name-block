FROM node:12-alpine

WORKDIR /usr/app
COPY lib lib
COPY package.json .
COPY VERSION .

RUN npm install

CMD ["npm", "start"]