# Up and running with Kafka and Go
This is a quick and dirty cradle to grave walk-through to get up and running with Go and Kafka producer/consumers. I assume that you have at least a basic understanding of Kafka as I will not be going over Kafka in large outside of what you see in the repo. We will utilize segment's [kafka-go client](https://github.com/segmentio/kafka-go) to create our producers and consumers. We will be using avro and confluents schema registry to encode/decode our data. To work with avro we will be utilizing [linkedin's goavro](https://github.com/linkedin/goavro) library and for our schema registry client will be the [landoop schema registry client](https://github.com/Landoop/schema-registry).  


## Setup

You'll need the following:

* [Go 1.7](https://golang.org/doc/install) or newer required.
* [go dep](https://github.com/golang/dep) dependency manager
* [docker](https://www.docker.com/) version 17.05 or newer
* [edward](https://github.com/yext/edward) tool for managing microservice instances
* [landoop's fast-data-dev](https://github.com/Landoop/fast-data-dev) we will use the docker image to create an environment that is both portable and reliable. As it is docker, the setup is painless.

**Note**: I will be speaking in the rest of the write up as an OSX user

### Setting up Landoop
If you want to bend fast-data-dev to your will, go straight to their [github repo](https://github.com/Landoop/fast-data-dev). 

**Note**: when accessing the UI, it is using the OSX hostname created by docker-machine, if you are using linux just replace it with localhost

#### LINUX 
As with most things software, linux is the easiest to setup. Just run:

```bash
docker run --rm --name ldp --net=host landoop/fast-data-dev
```

Pay attention to the `--name ldp`, as this is how you will enter the container to create topics or whatever else you want to run inside the container.

#### OSX
To run this properly in Mac you'll have a few hoops to jump through. First you'll need to run the following to create an environment for your docker containers to run in:

```bash
docker-machine create --driver virtualbox --virtualbox-memory 6000 landoop
```

You can verify that your new environment was successful by running:

```bash
docker-machine ls
```

You should see the name landoop.  Now you will need to get into that environment with the following command:

```bash
eval $(docker-machine env landoop)
```

Now you can run the following to start up your fast-data-dev container:

```bash
docker run --rm --name ldp -p 2181:2181 -p 3030:3030 -p 8081-8083:8081-8083 \
           -p 9581-9585:9581-9585 -p 9092:9092 -e ADV_HOST=192.168.99.100 \
           -e SAMPLEDATA=0 -e RUNTESTS=0 \
           landoop/fast-data-dev:latest
```
Pay attention to the `--name ldp`, as this is how you will enter the container to create topics or whatever else you want to run inside the container.

#### Windows

God help you....

#### Verify setup

open a browser and go to `http://192.168.99.100:3030`. You should see the landoop UI.

### Create the zip Topics

Make sure the fast-data-dev container is up and running. Eval into the correct docker-machine environment with the following:
 
```bash
eval $(docker-machine env landoop)
``` 
 
Now run the following to create the zip topic first:
```bash
docker exec -it ldp kafka-topics --create --zookeeper 192.168.99.100:2181 --topic zip --replication-factor 1 --partitions 3
```
the zip topic will have 3 partitions and a replication factor of 1

Now create the address enriched zipcode topic:
```bash
docker exec -it ldp kafka-topics --create --zookeeper 192.168.99.100:2181 --topic zip.address --replication-factor 1 --partitions 1 
```

Verify you have created the two new topics with the following:
```bash
docker exec -it ldp kafka-topics --list --zookeeper 192.168.99.100:2181
```

**NOTE**: You can also do all this with the bash script included, if on OSX, run the eval command above before the script

```bash
bash topic_setup.sh
```

### Add the necessary schemas to confluent's schema registry
This part is fairly straight forward. Jump back to your browser and go to [http://192.168.99.100:3030/schema-registry-ui/](http://192.168.99.100:3030/schema-registry-ui/). Click the NEW button in the upper right. 

One quick word before continuing, make sure to name your subjects with this format `topicname-value` or `topicname-key`. Now you can't say I didn't warn you...

Next create the schema for the zip topic messages as you see below

Subject Name = `zipcode-value`

Schema:
```json
{
  "type": "record",
  "name": "zipcode",
  "namespace": "com.landoop",
  "doc": "simple zipcode producer schema",
  "fields": [
    {
      "name": "zipcode",
      "type": "string"
    },
    {
      "name": "timestamp",
      "type": "long",
      "doc": "Time the message was encoded to avro (nanoseconds since epoch). May be used for ordering."
    }
  ]
}
```
click validate and viola, schema created

now create the schema for the enrichedZip topic messages

Subject Name = `zip_address-value`

Schema:
```json
{
  "type": "record",
  "name": "addressEnrichedZip",
  "namespace": "com.landoop",
  "doc": "Schema to define enriched zip data, enriched with addresses",
  "fields": [
    {
      "name": "zipcode",
      "type": "string"
    },
    {
      "name": "timestamp",
      "type": "long",
      "doc": "Time the message was encoded to avro (nanoseconds since epoch). May be used for ordering."
    },
    {
      "name": "addresses",
      "type": {
        "type": "array",
        "items": {
          "type": "record",
          "name": "address",
          "fields": [
            {
              "name": "zipcode",
              "type": "string"
            },
            {
              "name": "street",
              "type": "string"
            },
            {
              "name": "city",
              "type": "string"
            },
            {
              "name": "state",
              "type": "string"
            },
            {
              "name": "country",
              "type": "string"
            }
          ]
        }
      }
    }
  ]
}
```
again click validate and schema should have been created.

We are not going to create any keys for this demo, but you would follow the same syntax as you see above.

### Verify everything is working

Just run the following command from the root level of the repo:

```bash
edward start -t kafka-pipeline
```

you should see the producer logs and the consumer will start logging in like after 5-10 seconds or so. If that does not work, you should see some helpful logs to help you debug the issue.

**SETUP COMPLETE**
