# Running a Data Availability node

In order to run this service, the following components are needed:

1. Postgres DB
2. Ethereum node (L1)
3. The Data Availability node (DAN) itself

This three components can be run in many different ways, for reference, this guide will provide instructions using the following setup:

- L1 node is provided by a 3rd party, and so just the URL of it is needed
- The Postgres instance and the DAN run in docker containers, configured using `docker-compose`

Note that this is just an opinionated way of running things. It's possible for instance to run the DAN using the binnary, and Postgres in managed instance by a cloud provider (an many more configurations). 

```yml
version: "3.5"
networks:
  default:
    name: cdk-data-availability

services:

  cdk-data-availability:
    container_name: cdk-data-availability
    restart: unless-stopped
    depends_on:
      cdk-data-availability-db:
        condition: service_healthy
    image: hermeznetwork/cdk-data-availability:v0.0.2
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M
    ports:
      - 8444:8444
    volumes:
      - ./config.toml:/app/config.toml
      - ./private.keystore:/pk/test-member.keystore
    command:
      - "/bin/sh"
      - "-c"
      - "/app/cdk-data-availability run --cfg /app/config.toml"

  cdk-data-availability-db:
    container_name: cdk-data-availability-db
    restart: unless-stopped
    image: postgres:15
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -d $${POSTGRES_DB} -U $${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    ports:
      - 5434:5432
    environment:
      - POSTGRES_USER=committee_user            # CHANGE THIS: use your prefered user name
      - POSTGRES_PASSWORD=committee_password    # CHANGE THIS: use a safe and strong password
      - POSTGRES_DB=committee_db
    command:
      - "postgres"
      - "-N"
      - "500"
```

1. Copy/paste this into a new directory, name the file `docker-compose.yml`
2. In the same directory, create a file named `config.toml`, in that file, copy/paste the following:

```toml
PrivateKey = {Path = "/pk/test-member.keystore", Password = "testonly"} # CHANGE THIS (the password): according to the private key file password

[L1]
WsURL = "ws://URLofYourL1Node:8546"     # CHANGE THIS: use the URL of your L1 node
RpcURL = "http://URLofYourL1Node:8545"  # CHANGE THIS: use the URL of your L1 node
CDKValidiumAddress = "0x0DCd1Bf9A1b36cE34237eEaFef220932846BCD82"       # CHANGE THIS: Address of the Validium smart contract
DataCommitteeAddress = "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6"     # CHANGE THIS: Address of the data availability committee smart contract
Timeout = "3m"
RetryPeriod = "5s"
BlockBatchSize = 32

[Log]
Environment = "development" # "production" or "development"
Level = "debug"
Outputs = ["stderr"]

[DB]
User = "committee_user"             # CHANGE THIS: according to the POSTGRES_USER in docker-compose.yml
Password = "committee_password"     # CHANGE THIS: according to the POSTGRES_PASSWORD in docker-compose.yml
Name = "committee_db"
Host = "cdk-data-availability-db"
Port = "5432"
EnableLog = false
MaxConns = 200

[RPC]
Host = "0.0.0.0"
Port = 8444
ReadTimeout = "60s"
WriteTimeout = "60s"
MaxRequestsPerIPAndSecond = 500
```

3. Next step is to generate a file for the Ethereum private key of the committee member. Note that this private key should be representing one of the address of the committee. To generate the private key, run: `docker run -d -v .:/key hermeznetwork/zkevm-node /app/zkevm-node encryptKey --pk **** --pw **** -o /key`. Replace the **** for your actual private key and a password of your choice. After running the command, a file named `UTC--...` will be generated. Rename it to `private.keystore`
4. Change all the fields marked with `CHNAGE THIS` on both the `docker-compsoe.yml` and `config.toml`
5. Run it: `docker compose up -d`
6. Check the logs to see if everything is going fine: `docker compose logs`

Note that the DAN endpoint (in this example using the port 8444) should be reachable in the URL indicated on the data availability smart contract