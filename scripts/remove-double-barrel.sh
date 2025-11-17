#!/bin/bash

# Script to remove double-barrel (hyphenated) animal names from noun.txt

if [ $# -lt 1 ]; then
    echo "Usage: $0 <noun.txt file>"
    echo "Example: $0 data/animals/noun.txt"
    exit 1
fi

FILE="$1"

if [ ! -f "$FILE" ]; then
    echo "Error: File '$FILE' not found"
    exit 1
fi

# Create a temporary file
TMPFILE=$(mktemp)

# Count removed items
REMOVED=0

# Process each line
while IFS= read -r line || [ -n "$line" ]; do
    # Trim whitespace (handle quotes properly)
    trimmed=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    # Check if line contains a hyphen (double-barrel name)
    if [[ "$trimmed" == *"-"* ]]; then
        echo "Removed: $trimmed"
        ((REMOVED++))
    else
        # Keep the line (including empty lines)
        echo "$line" >> "$TMPFILE"
    fi
done < "$FILE"

# Move temp file to original file
mv "$TMPFILE" "$FILE"

# Count remaining names (non-empty lines)
KEPT=$(grep -v '^[[:space:]]*$' "$FILE" | wc -l | xargs)

echo ""
echo "Done! Removed $REMOVED double-barrel names from $FILE"
echo "Kept $KEPT single-word names"

