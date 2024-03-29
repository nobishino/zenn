
git diff --name-only | grep "gosample/main.go"
if [ $? -eq 0 ]; then
  echo "gosample/main.go has been changed. construct the commit message..."
  echo $1 > tmp
  echo "" >> tmp
  pushd gosample
  echo "$(gp share main.go go.mod go.sum)?v=gotip"  >> ../tmp
  popd 
  echo "" >> tmp
  echo  "\`\`\`go" >> tmp
  cat gosample/main.go >> tmp
  echo  "\`\`\`" >> tmp
  git commit -am "$(cat tmp)"
fi