## Udpbeat

ELK beat that collects the structured inputs via UDP and emits them to ELK

**Status:** expermental, use at your own risk

### Setup

Add trace's document template to your ElasticSearch cluster**

```shell
curl -XPUT 'http://localhost:9200/_template/trace' -d@udbbeat/template.json
```

**Start udpbeat UDP logs collector and emitter**

```shell
go install github.com/gravitational/udpbeat
udpbeat
```

Start emitting logs to UDP socket via `github.com/gravitational/trace` UDP hook and enjoy structured logs!
