# Use an official Python runtime as a parent image
FROM python:3.8-slim

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Install any needed packages specified in requirements.txt
RUN apt-get -y update \
    && apt-get -y install git \
    && pip install --no-cache-dir -r /app/requirements.txt \
    && apt-get remove -y git \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Run edge_client.py when the container launches
CMD ["python", "-u" "/app/edge_client.py"]