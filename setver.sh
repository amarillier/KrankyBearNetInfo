#! /bin/sh

if [ $# -ge 1 ]
then
    ver=$1
else
    echo "Enter a version number"
    cur=$(cat main.go | grep -i "appVersion" | grep "=" | awk '{print $3}' | tr -d '"')
    echo "    current: $cur"
    read ver
    if [ -z "$ver" ]
    then
        echo "Enter a version!"
        echo "No version change detected, continuing to allow compile to continue"
        exit
    else
        echo "Version: $ver"
        # exit
    fi
fi

echo "version: $ver"
echo "main.go"
sed -i '' "s/appVersion = \".*\"/appVersion = \"$ver\"/" main.go

if [ -f "FyneApp.toml" ]
then
    echo "FyneApp.toml found"
    sed -i '' "s/Version = \".*\"/Version = \"$ver\"/" FyneApp.toml
else
    echo "FyneApp.toml not found"
fi

echo "Inno Setup winres/winres.json"
sed -i '' "s/file_version\":.*/file_version\": \"$ver\",/" ./winres/winres.json
sed -i '' "s/product_version\":.*/product_version\": \"$ver\"/" ./winres/winres.json
sed -i '' "s/FileVersion\":.*/FileVersion\": \"$ver\",/" ./winres/winres.json
sed -i '' "s/ProductVersion\":.*/ProductVersion\": \"$ver\",/" ./winres/winres.json

# "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
