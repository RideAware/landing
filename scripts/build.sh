#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
IMAGE_NAME="rideaware-landing"
IMAGE_TAG="latest"
NO_CACHE=false
RUN_CONTAINER=false
CONTAINER_NAME="rideaware-landing"

# Help function
show_help() {
	cat << EOF
Usage: $0 [OPTIONS]

OPTIONS:
  -t, --tag TAG           Image tag (default: latest)
  -n, --name NAME         Image name (default: rideaware-landing)
  -r, --run               Run container after build
  -c, --container NAME    Container name when running (default: rideaware-landing)
  --no-cache              Build without cache
  -h, --help              Show this help message

EXAMPLES:
  $0                              # Build as rideaware:latest
  $0 -t v1.0                      # Build as rideaware:v1.0
  $0 -t dev --run                 # Build and run
  $0 --no-cache -t prod           # Build without cache as rideaware:prod

EOF
	exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
	case $1 in
		-t|--tag)
			IMAGE_TAG="$2"
			shift 2
			;;
		-n|--name)
			IMAGE_NAME="$2"
			shift 2
			;;
		-r|--run)
			RUN_CONTAINER=true
			shift
			;;
		-c|--container)
			CONTAINER_NAME="$2"
			shift 2
			;;
		--no-cache)
			NO_CACHE=true
			shift
			;;
		-h|--help)
			show_help
			;;
		*)
			echo -e "${RED}Unknown option: $1${NC}"
			show_help
			;;
	esac
done

FULL_IMAGE="$IMAGE_NAME:$IMAGE_TAG"
BUILD_ARGS=""

if [ "$NO_CACHE" = true ]; then
	BUILD_ARGS="--no-cache"
fi

# Function to stop and remove container
cleanup_container() {
	local name=$1

	if podman ps -a --format "{{.Names}}" | grep -q "^${name}\$"; then
		echo -e "${YELLOW}Removing existing container: $name${NC}"

		# Stop if running
		if podman ps --format "{{.Names}}" | grep -q "^${name}\$"; then
			echo "  Stopping container..."
			podman kill "$name" 2>/dev/null || true
		fi

		# Remove
		echo "  Removing container..."
		if podman rm "$name" 2>/dev/null; then
			echo -e "${GREEN}  ✓ Container removed${NC}"
		else
			echo -e "${RED}  ✗ Failed to remove container${NC}"
			return 1
		fi
	fi
	return 0
}

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        Building Podman Image           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo -e "${YELLOW}Image: $FULL_IMAGE${NC}"
echo ""

if ! podman build $BUILD_ARGS -f Containerfile -t "$FULL_IMAGE" .; then
	echo -e "${RED}✗ Build failed${NC}"
	exit 1
fi

echo -e "${GREEN}✓ Image built successfully${NC}"
echo ""

# Show image info
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        Image Details                   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
podman images "$IMAGE_NAME:$IMAGE_TAG" \
	--format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.Created}}"
echo ""

if [ "$RUN_CONTAINER" = true ]; then
	echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
	echo -e "${BLUE}║      Starting Container               ║${NC}"
	echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"

	# Cleanup existing container
	if ! cleanup_container "$CONTAINER_NAME"; then
		echo -e "${RED}✗ Failed to clean up existing container${NC}"
		exit 1
	fi

	echo ""
	echo "Starting new container: $CONTAINER_NAME"

	if podman run -d \
		--name "$CONTAINER_NAME" \
		-p 5000:5000 \
		--env-file .env \
		"$FULL_IMAGE"; then
		echo -e "${GREEN}✓ Container running: $CONTAINER_NAME${NC}"
		echo ""

		# Wait for startup
		sleep 2

		echo -e "${YELLOW}Container logs:${NC}"
		podman logs "$CONTAINER_NAME"
		echo ""

		echo -e "${GREEN}Site available at: http://localhost:5000${NC}"
		echo -e "${YELLOW}To view logs: podman logs -f $CONTAINER_NAME${NC}"
		echo -e "${YELLOW}To stop: podman kill $CONTAINER_NAME${NC}"
	else
		echo -e "${RED}✗ Failed to start container${NC}"
		exit 1
	fi
else
	echo -e "${YELLOW}To run the container:${NC}"
	echo "  podman run -d --name $CONTAINER_NAME -p 5000:5000 --env-file .env $FULL_IMAGE"
	echo ""
	echo -e "${YELLOW}Or use this script with --run:${NC}"
	echo "  $0 -t $IMAGE_TAG --run"
fi

echo ""
echo -e "${GREEN}✓ Done!${NC}"