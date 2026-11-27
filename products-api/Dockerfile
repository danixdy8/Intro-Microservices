FROM node:22-slim

WORKDIR /app

RUN apt-get update && apt-get install -y python3 make g++ && rm -rf /var/lib/apt/lists/*

COPY package.json ./
RUN npm install --omit=dev

COPY src ./src

RUN mkdir -p /data
ENV DB_PATH=/data/products.db
ENV PORT=8003

EXPOSE 8003

CMD ["node", "src/index.js"]
