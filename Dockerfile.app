# Use an official lightweight Node.js image as a base image
FROM node:14-alpine

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy the package.json and package-lock.json files to the working directory
COPY package*.json ./

# Install any dependencies if needed
RUN npm install

# Copy the contents of the app directory to the working directory
COPY app/ .

# Expose the port on which your application will run
EXPOSE 8080

# Define the command to run your application
CMD [ "npm", "start" ]
