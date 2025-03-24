ID = $1

if [ -z "$ID" ]; then
    echo "Usage: ./run.sh <ID>"
    exit 1
fi

trap 'echo "Ignoring Ctrl+C.";' SIGINT

while true; do
    echo "Building the project."
    go build -o elevator main.go || { echo "Build failed. Retrying."; sleep 1; continue }

    echo "Starting elevator with ID=$ID..."

    echo "Program crashed or terminal closed. Restarting in a new window."

    sleep 1
    gnome-terminal -- bash -c "cd $(pwd); ./run.sh $ID; exec bash"
    exit
done