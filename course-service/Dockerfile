FROM node:20-alpine

# Set working directory
WORKDIR /app

# Copy dependency info
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Run app
CMD ["node", "src/app.js"]