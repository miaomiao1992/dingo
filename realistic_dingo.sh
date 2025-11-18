#!/bin/bash

# Realistic ASCII Dingo Animation
# Authentic Australian Dingo
# Press Ctrl+C to exit

clear

# Walking animation frames
frame1='
          /|        /|
         / |       / |
         | |       | |       🦮 AUSTRALIAN DINGO
          \|        \|
        _."          "._
      ."  _    ..jijj    _  ".
     /   //\    /^  ^\   /\\  \\
    |   /  ||   \/>  <   ||  |  |
    |  /   \|    |  / \  |/   |  |
    | |     |    |  ( )      |  |  |
    | |     |    |   ) |     |  |  |
    | |     |    |  /  |     |  |  |
    | |     |    |/   /      |  |  |
    | |     |         /    |  |  |
    |_|     |_       /      _|_ |  |
    /|       |\     /|\   /|   |  |
   / |       | \   / | \  |   |  /
  /_/         \ \_/  \_/   |_ /__/
'

frame2='
           /|        /|
          / |       /  |       🦮 AUSTRALIAN DINGO
           \|       \   |
         _."          "._
       ."  _     ^     _    ".
      /  _    \\    //  _    \\
     |  _     | |  | |     _  |
    | _     | |   | |    _   _|
    |  ^     | |   | |  ^  ^   |
   _|   |    | |   | |       _|    Front paw up
  /     |    /  |   |  |     /
 |     ^    |     \  |  |    /
 |     |    |      \_|  |   /
 |     |    |           |  |
'

frame3='
          /|        /|
         / |       /  |       🦮 AUSTRALIAN DINGO
          \|       \   |
         _."             "._
       ."  _      ^      ".
      /  _    \\\\_    //  _
     |  _    /  | |     _  _|
    | ^     |   | |  ^     ^
    | |     |   | |   |   | |
    | |     |   | |   |   | |
    | |     |   | |   |   | |
    | |     |   | |   |   | |
    |_|      \_/ \_/    _|
'

frame4='
          /|        /|
         / |       /  |       🦮 AUSTRALIAN DINGO
           \|          \   |
         _."             "._
       ."  _      ^           ".
      /  _    \\\\    //  _     \\
     | ^     | | |  | |   ^  |    Back paw up
    | |     | | |  | |   |  _|
    | |     | | |  | |   |  ^
    | |     | | |  | |   | |
    | |     | | |  | |   | |
    | |     | | |  | |   | |
    |_|      \_/   \_/   |_|
'

frames=("$frame1" "$frame2" "$frame3" "$frame4")

# Function to display walking animation
animate() {
    local count=0
    local max_count=${#frames[@]}

    while true; do
        clear
        echo "${frames[$count]}"

        # Show different descriptions
        case $count in
            0)
                echo ""
                echo "    🦘 Australian Dingo - The Apex Predator"
                ;;
            1)
                echo ""
                echo "    🦘 Wild Dog - Nocturnal Hunter"
                ;;
            2)
                echo ""
                echo "    🦘 Native Australian - Solitary & Majestic"
                ;;
            3)
                echo ""
                echo "    🦘 Australian Dingo - Apex Predator"
                ;;

        esac

        echo ""
        echo "    Press Ctrl+C to exit"
        echo ""
        echo "    Frame: $((count + 1))/$max_count)"

        sleep 0.7
        count=$(( (count + 1) % max_count ))
    done
}

# Trap Ctrl+C for graceful exit
trap 'clear; echo ""; echo "  🎯 Thanks for watching the Dingo!"; echo "  🦮 Australia's Wild Dog"; echo ""; exit 0' INT

animate