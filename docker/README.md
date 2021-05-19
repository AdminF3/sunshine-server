# For development purpose

## Run sunshine

- Run `mkdir $HOME/sunshine-uploads` in current docker directory ( this should be executed only once not everytime when you try to run the container )
- Run `docker-compose build --no-cache dev-sunshine` in current docker directory
- Run `docker-compose up -d` in current docker directory
- Your project is listening on port 8001 on ur local machine , Voila !

## Update of sunshine image

- Run `docker-compose build --no-cache dev-sunshine` in current docker directory
- Run `docker rm dev-sunshine -f` in current docker directory
- Run `docker-compose up -d` in current docker directory
