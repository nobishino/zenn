
git diff --name-only | grep "gosample/main.go"
if [ $? -eq 0 ]; then
  echo "gosample/main.go has been changed. construct the commit message..."
  echo $1 > tmp
  echo "" >> tmp
  gp share ./gosample/main.go ./gosample/go.mod >> tmp
  echo "" >> tmp
  echo  "\`\`\`go" >> tmp
  cat gosample/main.go >> tmp
  echo  "\`\`\`" >> tmp
  git commit -am "$(cat tmp)"
fi