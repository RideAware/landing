#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

IMAGE_NAME="rideaware-landing"
IMAGE_TAG="latest"
CONTAINER_NAME=""
RUN_AFTER=false
NO_CACHE=""

show_help() {
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -t, --tag TAG         Image tag (default: latest)"
    echo "  -n, --name NAME       Image name (default: rideaware-landing)"
    echo "  -r, --run             Run container after build"
    echo "  -c, --container NAME  Container name for running"
    echo "      --no-cache        Build without cache"
    echo "  -h, --help            Show this help"
    echo ""
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag) IMAGE_TAG="$2"; shift 2 ;;
        -n|--name) IMAGE_NAME="$2"; shift 2 ;;
        -r|--run) RUN_AFTER=true; shift ;;
        -c|--container) CONTAINER_NAME="$2"; shift 2 ;;
        --no-cache) NO_CACHE="--no-cache"; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) echo "Unknown option: $1"; show_help; exit 1 ;;
    esac
done

FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

echo -e "${CYAN}Building ${FULL_IMAGE}...${NC}"
podman build ${NO_CACHE} -t "${FULL_IMAGE}" -f Containerfile . || {
    echo -e "${RED}Build failed${NC}"
    exit 1
}

echo -e "${GREEN}Build successful: ${FULL_IMAGE}${NC}"
podman images "${IMAGE_NAME}"

if [ "$RUN_AFTER" = true ]; then
    NAME_FLAG=""
    if [ -n "$CONTAINER_NAME" ]; then
        # Stop and remove existing container
        podman kill "$CONTAINER_NAME" 2>/dev/null
        podman rm "$CONTAINER_NAME" 2>/dev/null
        NAME_FLAG="--name ${CONTAINER_NAME}"
    fi

    echo -e "${CYAN}Starting container...${NC}"
    podman run -d ${NAME_FLAG} -p 5000:5000 --env-file .env "${FULL_IMAGE}"

    echo -e "${GREEN}Container started on http://localhost:5000${NC}"
    if [ -n "$CONTAINER_NAME" ]; then
        podman logs "$CONTAINER_NAME"
    fi
fi
