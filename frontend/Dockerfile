FROM ubuntu

# Install
RUN apt-get update
RUN apt-get install -y nginx

# Config
RUN echo "daemon off;" >> /etc/nginx/nginx.conf
ADD ./default /etc/nginx/sites-available/default

# HTML files
RUN mkdir -p /var/www/html
ADD . /var/www/html

# Port exposure
EXPOSE 80 

# Run it
CMD nginx
