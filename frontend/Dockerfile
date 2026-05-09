FROM node:20-alpine

WORKDIR /app

# Bağımlılıkları kopyala ve yükle
COPY package.json ./
RUN npm install

# Kaynak kodunu kopyala
COPY . .

# Geliştirme ortamı için Vite'i dışarıya açarak çalıştır
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
