# example Dockerfile for https://docs.docker.com/engine/examples/postgresql_service/

FROM gedilabs/base:v1 

# Add the PostgreSQL PGP key to verify their Debian packages.
# It should be the same key as https://www.postgresql.org/media/keys/ACCC4CF8.asc
#RUN apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8

# Add PostgreSQL's repository. It contains the most recent stable release
#  of PostgreSQL.
#RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ precise-pgdg main" > /etc/apt/sources.list.d/pgdg.list

# Install ``software-properties-common`` and PostgreSQL 11
#  There are some warnings (in red) that show up during the build. You can hide
#  them by prefixing each apt-get statement with DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y software-properties-common postgresql-11 postgresql-client-11 postgresql-contrib-11

# Note: The official Debian and Ubuntu images automatically ``apt-get clean``
# after each ``apt-get``

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-11`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN service postgresql start &&\
    psql --command "CREATE USER rodinia WITH SUPERUSER PASSWORD 'rodinia';" &&\
    createdb -O rodinia rodinia
ENV PGPASSWORD=rodinia
RUN service postgresql start &&\
    psql -h localhost -U rodinia -d rodinia --command "CREATE TABLE actors (index serial, \ 
                                         id text PRIMARY KEY, \
                                         addr text)" &&\
    psql -h localhost -U rodinia -d rodinia --command "CREATE TABLE dht (id text PRIMARY KEY, \
                                       keys text[], \
                                       successor text)" 

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all rodinia all trust" >> /etc/postgresql/11/main/pg_hba.conf
RUN echo "local all rodinia trust" >> /etc/postgresql/11/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/11/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/11/main/postgresql.conf

# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
# removed "/var/lib/postgresql" from this list
VOLUME  ["/etc/postgresql", "/var/log/postgresql"]

# Set up satellite microservice 

# Copy application package
#(note: we are bulding from top-level in project directory

COPY . /gedilabs 
COPY ./app/satellite/entrypoint.sh /usr/bin/entrypoint.sh 

USER root

RUN chmod u+x /usr/bin/entrypoint.sh 

ENTRYPOINT ["entrypoint.sh"]

WORKDIR /gedilabs/app/satellite

CMD ["go", "run", "main.go"]

