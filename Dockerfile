FROM node:12-alpine

WORKDIR /usr/app
COPY lib lib
COPY package.json .
COPY VERSION .
COPY data .

RUN npm install

CMD ["npm", "start"]