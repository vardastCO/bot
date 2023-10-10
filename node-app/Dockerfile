FROM ghcr.io/puppeteer/puppeteer:19.7.2
WORKDIR /usr/src/app

# We don't need the standalone Chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD true \ 
   PUPPETEER_EXECUTABLE_PATH = /usr/bin/google-chrome-stable


COPY . .

RUN npm install

# Expose port 3002 for your Node.js application.
EXPOSE 3002

# Start your Node.js application.
CMD ["node", "scrape.js"]
