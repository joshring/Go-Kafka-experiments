# Go-Kafka-experiments

## To run: 
1. docker compose up

### Notes:
- All containers compile as they are run, and recompile on changes, no need to build them separately.
- JSON dump outputs are stored in ```output_dumps```.
- Simple API reads the outputs to find the last recorded date and outputs that back to the user.
- Using a monorepo with Go workspaces and a `go.work` file, docker builds contain the whole monorepo context

### Development environment:
- Ubuntu Linux 22.04 LTS
- Docker version 24.0.5, build ced0996
- Docker Compose version v2.20.2




