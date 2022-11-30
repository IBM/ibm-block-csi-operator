docker build -f build/ci/Dockerfile.unittest -t volume-group-operator-unittestss .
docker run --rm -t volume-group-operator-unittestss
