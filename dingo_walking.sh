#!/bin/bash

# Walking Dingo ASCII Animation
# Press Ctrl+C to exit

clear

# Animation frame counter
frame=0

# Function to display a frame
show_frame() {
    clear

    case $1 in
        0)
            cat << 'EOF'

                    🐕 AUSTRALIAN DINGO 🐕
          |        /|\        |
         /|\       / | \       |\
        / | \      | | |      / | \
        | | |      | | |      | | |
        | | |      | | |      | | |
      _.| | |      | | |      | | | _.
    .'  | | |  ..jijj  |      | | |  '.
   /   /| | |    /\      |      | | |\   \
  |   / | | |   |  |     |      | | | \   |
  |  /  | | |   |  |     |      | | |  \  |
  | |   | | |   |  |     |      | | |   | |
  | |   | | |  /____\    |      | | |   | |
  | |   | | |  |    |    |      | | |   | |
  | |   | | |  |    |    |      | | |   | |
  | |   | | |  |    |    |      | | |   | |
  |_|   |_| |__|    |____|      |_| |   _|
EOF
            ;;
        1)
            cat << 'EOF'

                    🐕 AUSTRALIAN DINGO 🐕
           |        /|\        |
          /|\       / | \       |\
         / | \      | | |      / | \
         | | |      | | |      | | |
         | | |      | | |      | | |
       _.| | |      | | |      | | | _.
     .'  | | |   ^       |      | | |  '.
    /   /| | |    \     ^ |      | | |\   \
   |   / | | |     \      |      | | | \   |
   |  /  | | |      \     |      | | |  \  |
   | |   | | |       \    |      | | |   | |
   | |   | | |    [front paw up] | | |   | |
   | |   | | |      /    |      | | |   | |
   | |   | | |     /     |      | | |   | |
   |_|   |_| |____/      |______|_| |   _|
EOF
            ;;
        2)
            cat << 'EOF'

                    🐕 AUSTRALIAN DINGO 🐕
          |        /|\        |
         /|\       / | \       |\
        / | \      | | |      / | \
        | | |      | | |      | | |
        | | |      | | |      | | |
      _.| | |      | | |      | | | _.
    .'  | | |      | | |      | | |  '.
   /   /| | |      | | |      | | |\   \
  |   / | | |   \\\\|//      | | | \   |
  |  /  | | |    \\\\|//      | | |  \  |
  | |   | | |      ||         | | |   | |
  | |   | | |      ||         | | |   | |
  | |   | | |      ||         | | |   | |
  | |   | | |      ||         | | |   | |
  |_|   |_| |______||_________|_| |   _|
EOF
            ;;
        3)
            cat << 'EOF'

                    🐕 AUSTRALIAN DINGO 🐕
          |        /|\        |
         /|\       / | \       |\
        / | \      | | |      / | \
        | | |      | | |      | | |
        | | |      | | |      | | |
      _.| | |      | | |      | | | _.
    .'  | | |  ^       |      | | |  '.
   /   /| | |    \     ^ |      | | |\   \
  |   / | | |     \      |      | | | \   |
  |  /  | | |      \     |      | | |  \  |
  | |   | | |       \    |      | | |   | |
  | |   | | |   [back paw up]   | | |   | |
  | |   | | |      /    |      | | |   | |
  | |   | | |     /     |      | | |   | |
  |_|   |_| |____/      |______|_| |   _|
EOF
            ;;
    esac

    echo ""
    echo "    Frame: $((frame + 1))/4 - Walking Animation"
    echo ""
    echo "    🐺 Australia's Wild Dog 🐺"
    echo "    Press Ctrl+C to exit"
}

# Main animation loop
while true; do
    show_frame $frame
    sleep 0.8
    frame=$(( (frame + 1) % 4 ))
done

# Trap Ctrl+C
trap 'clear; echo ""; echo "  🎯 Thanks for watching the Dingo!"; echo "  🦮 Australia's Wild Dog"; echo ""; exit 0' INT