#!/bin/bash
# Template: set-git-cred.sh
# Usage: Edit the 'project' variable for your repo before running.

project=boki-webapps

# Display project info
echo "-----------------------------------------------------------------------------"
echo "this is project https://github.com/bojanstrkovski-21/$project"
echo "-----------------------------------------------------------------------------"

# Set global git config (edit as needed)
git config --global pull.rebase false
git config --global user.name "bojanstrkovski-21"
git config --global user.email "bojanstrkovski.21@gmail.com"
sudo git config --system core.editor nvim
git config --global push.default simple

git remote set-url origin git@github.com:bojanstrkovski-21/$project

echo "Everything set"

echo "################################################################"
echo "###################    T H E   E N D      ######################"
echo "################################################################"
