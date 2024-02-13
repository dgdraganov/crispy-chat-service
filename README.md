
# Crispy-Chat-Serice

A simple chat server with example clients to publish and read messages.

## How to run?

The below examples illustrate how the functionality can be tested. 

In the terminal navigate to the project root directory.

The following make command will run the chat service and the redis client in detached mode:
```
    make server
```

Once the service is up the following command will start two chat bots that will start publishing messages to the chat room every few seconds:
```
    make bots
```

Finally the `client` make command will start a client that will consume all existing messages and will wait for new ones:
```
    make client
```

When finished the below command will stop all docker services related to the chat service:
```
    make decompose
```

## Private key

In the provided `docker-compose.yaml` file there is an example env variable containing base64 encoded private key. It has been generated with the following command:

```
    openssl ecparam -name secp521r1 -genkey -noout -out private.pem
```







