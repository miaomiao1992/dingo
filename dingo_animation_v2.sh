#!/bin/bash

# ASCII Dingo Animation v2.0
# Press Ctrl+C to exit

clear

# Animation frames - walking dingo
frames=(
"
                    __
                 .-\"  \"-.
                /        \
               |  _  _  |   Hello! I'm a Dingo!
               | | | | |
               | | | | |
              _| | | | |_
            .'  | | | |  '.
           /     | | |     \
          /      | | |      \
         /       | | |       \
        /        | | |        \
       /         | | |         \
      /          | | |          \
     /           | | |           \
    /_            | |            _\
      '--.      .-| |-.      .--'
           '.  .'  |  '.  .'
             '.    |    .'
               '-. | .-'
                  'V'
"
"
                      __
                   .-\"  \"-.
                  /        \
                 |  _  _  |   Hello! I'm a Dingo!
                 | | | | |
                 | | | | |
                _| | | | |_
              .'  | | | |  '.
               \     | |     /
                \      |      /
                 \       |       /
                  \       |       /
                   \       |       /
                    \       |       /
                     \_       _/
                       '-. .'
                          V
"
"
                   __
                .-\"  \"-.
               /        \
              |  _  _  |   Hello! I'm a Dingo!
              | | | | |
              | | | | |
             _| | | | |_
           .'  | | | |  '.
            \     | |     /
             \      |      /
              \       |       /
               \       |       /
                \       |       /
                 \       |       /
                  \_       _/
                    '-. .'
                       V
"
)

# Function to display animation
animate() {
    local count=0
    while true; do
        clear
        echo "${frames[$count]}"
        echo ""
        echo "🦮 Dingo Animation - Press Ctrl+C to exit"
        sleep 0.8
        count=$(( (count + 1) % ${#frames[@]} ))
    done
}

# Trap Ctrl+C
trap 'clear; echo ""; echo "Thanks for watching the Dingo! 🐺"; echo ""; exit 0' INT

animate