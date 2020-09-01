Reporting-app Setup

# Nodejs Reporting app
This utility can be used to host allure reports on the server. We have a support for api which will take input from the zip allure-results folder and will add results to the allure reporter running on the machine.
 
## Setup
This service runs on go. Please find steps to setting up the reporting-app on ubuntu 16.04 server.
 
- Install go
 - sudo apt-get update
 - sudo apt-get -y upgrade
 - cd /tmp
 - wget https://dl.google.com/go/go1.14.linux-amd64.tar.gz
 - sudo tar -xvf go1.14.linux-amd64.tar.gz
 - sudo mv go /usr/local
 
- Setup go Environment
 - export GOROOT=/usr/local/go
 - export GOPATH=$HOME/go
 - export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
 - source ~/.profile
 
- Verify Installation
 - go version
 
- Bring up Allure report on same ubuntu server:
   - sudo apt-get update
   - sudo apt-get install nodejs
   - sudo apt-get install npm
   - nodejs -v (Check node js is installed)
   - sudo apt install default-jre (To bring Allure reporter up it requires Java to be installed)
   - java -version (Check java is installed)
   - npm install -g allure-commandline --save-dev (allure reporter cli is required for the reporting-app)
   - allure (Check allure is installed)
 
- Bring the allure report server up:
   - allure serve --port #publicPortNumber (This command can be run in background on ubuntu server on tmux session)
 
- Checkout the code and build the project:
 - git clone https://github.com/ksfshah3/reporting-app.git
 
 - cd reporting-app/
 
By default the reporting app will run 8010 port, Please change in config.json the desired port where you want to run the report.
Also we can change username and password required for adding results to server in config.json
 
 - go run main.go >> reporting-app.logs
 
The application should start on the mentioned port.
 
- Curl used to add results:
  - curl -L -X POST 'http://localhost:8010/v1/allure/addresults/user=admin/password=admin123' -F 'file=@allure-results-zip.tar.gz'
 
The curl expects Zipped file of allure-results from local or Jenkins job run which would be pushed to the server. To get a reference on how to do that for Mac and Windows machines. Please check below files for example:
https://github.com/ksfshah3/reporting-app/blob/master/report-to-server.bat
https://github.com/ksfshah3/reporting-app/blob/master/repoort-to-server.sh
