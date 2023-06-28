# consent-to-fhir
![go](https://github.com/diz-unimr/consent-to-fhir/actions/workflows/build.yml/badge.svg) ![docker](https://github.com/diz-unimr/consent-to-fhir/actions/workflows/release.yml/badge.svg) [![codecov](https://codecov.io/gh/diz-unimr/consent-to-fhir/branch/main/graph/badge.svg?token=D66XMZ5ALR)](https://codecov.io/gh/diz-unimr/consent-to-fhir)
> Kafka processor to map gICS consent data to FHIR 🔥

This project consists of a Kafka consumer/producer to read gICS notification data from an input topic
and send the mapped FHIR Consent resources to an output topic.

The mapper currently relies on gICS itself to do the FHIR mapping and only harmonizes resources 
references and identifiers as well as profile specific data.
It currently supports the MII Consent module profile, only. 

Mapping directly from notification data might be implemented in the future.

## Notifications

This processor relies on notification data from gICS (json) to be consumed from a Kafka topic.
In order to provide this data, the [gics-to-kafka](https://github.com/diz-unimr/gics-to-kafka.git) 
producer can be used.

## Mapping

### TTT-FHIR Gateway

giCS supports mapping consent data to FHIR via the [TTP-FHIR Gateway](https://www.ths-greifswald.de/wp-content/uploads/tools/fhirgw/ig/2023-1-0/ImplementationGuide-markdown-Einwilligungsmanagement.html).

The `consent-to-fhir` mapper uses the [$currentPolicyStatesForPerson](https://www.ths-greifswald.de/wp-content/uploads/tools/fhirgw/ig/2023-1-0/ImplementationGuide-markdown-Einwilligungsmanagement-Operations-currentPolicyStatesForPerson.html) 
operation to get current policy states according to the input notification data.
This data is then mapped to the MII FHIR Consent profile and references and identifiers are set to local systems.  

### Supported consents and profiles

Currently, only the MII Broad consent (version 1.6.d) and the FHIR Consent module profile is supported.

## Configuration properties

| Name                             | Default                                                    | Description                                 |
|----------------------------------|------------------------------------------------------------|---------------------------------------------|
| `app.name`                       | consent-to-fhir                                            | Application name                            |
| `app.log-level`                  | info                                                       | Log level (error,warn,info,debug,trace)     |
| `app.mapper.consent-system`      | https://fhir.diz.uni-marburg.de/sid/consent-id             | Consent FHIR identifier system              |
| `app.mapper.patient-system`      | https://fhir.diz.uni-marburg.de/sid/patient-id             | Patient FHIR identifier system              |
| `app.mapper.domain-system`       | https://fhir.diz.uni-marburg.de/fhir/sid/consent-domain-id | Consent domain FHIR identifier system       |
| `kafka.bootstrap-servers`        | localhost:9092                                             | Kafka brokers                               |
| `kafka.security-protocol`        | ssl                                                        | Kafka communication protocol                |
| `kafka.ssl.ca-location`          | /app/cert/kafka-ca.pem                                     | Kafka CA certificate location               |
| `kafka.ssl.certificate-location` | /app/cert/app-cert.pem                                     | Client certificate location                 |
| `kafka.ssl.key-location`         | /app/cert/app-key.pem                                      | Client key location                         |
| `kafka.ssl.key-password`         | private-key-password                                       | Client key password                         |
| `kafka.input-topic`              |                                                            | Notification input topic                    |
| `kafka.output-topic`             |                                                            | Consent FHIR output topic                   |
| `kafka.num-consumers`            | 1                                                          | Number of concurrent Kafka consumer threads |
| `gics.signer-id`                 | Patienten-ID                                               | Target consent signerId                     |
| `gics.fhir.base`                 |                                                            | TTP-FHIR base url                           |
| `gics.fhir.auth.user`            |                                                            | TTP-FHIR Basic auth user                    |
| `gics.fhir.auth.password`        |                                                            | TTP-FHIR Basic auth password                |


### Environment variables

Override configuration properties by providing environment variables with their respective names.
Upper case env variables are supported as well as underscores (`_`) instead of `.` and `-`.


# Deployment

Example via `docker compose`:
```yml
consent-to-fhir:
    image: ghcr.io/diz-unimr/consent-to-fhir:latest
    restart: unless-stopped
    environment:
      APP_NAME: consent-to-fhir
      APP_LOG_LEVEL: info
      KAFKA_BOOTSTRAP_SERVERS: kafka:19092
      KAFKA_SECURITY_PROTOCOL: SSL
      KAFKA_INPUT_TOPIC: consent-json-idat
      KAFKA_OUTPUT_TOPIC: consent-fhir-idat
      GICS_SIGNER_ID: Patienten-ID
      GICS_FHIR_BASE: https://gics.local/ttp-fhir/fhir/gics/
      GICS_FHIR_AUTH_USER: test
      GICS_FHIR_AUTH_PASSWORD: test
    volumes:
     - ./cert/ca-cert:/app/cert/kafka-ca.pem:ro
     - ./cert/consent-to-fhir.pem:/app/cert/app-cert.pem:ro
     - ./cert/consent-to-fhir.key:/app/cert/app-key.pem:ro
```

# License

[AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.en.html)