# golang-github-enterprise-runners

So, you want to automatically register GitHub Action Self-Hosted Runners for your Orginizatiton. You've probably looked at the docs and thought "why did they make it so difficult?" Well, here's a fully working proof-of-concept using Go that will generate the various tokens and use them to call the APIs needed to (finally) get a Runner registration token. 

### !!! NOTE !!! 
This code logs secret values and should ***NOT*** be used in a production environment! It is provided as a proof-of-concept so you can see the various steps needed to get a registration token. 

## Prerequisites
1. You've created a [GitHub application](https://docs.github.com/en/developers/apps/building-github-apps/creating-a-github-app) and given it `Read & Write` permissions for `Organization > Self-hosted runners`. 
2. You've [installed the app](https://docs.github.com/en/enterprise-server@3.4/developers/apps/managing-github-apps/installing-github-apps) to the orginization for which you want to manage runners. 
3. You've [generated a private key](https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#generating-a-private-key) for the app and have a copy of the file. 
4. You've noted the [App ID](https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app) and its [Installation ID](https://docs.github.com/en/rest/orgs/orgs#list-app-installations-for-an-organization). 

Now you can update the `const` values in the `main.go` file and run it to get your runner registration tokens. 
