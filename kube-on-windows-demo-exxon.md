## Environment setup
Use testkit with Azure to provision the nodes with the following credentials:

```bash
export TESTKIT_AZURE_DISK_SIZE_GB=128
export TESTKIT_AZURE_LOCATION=westus2
export TESTKIT_AZURE_MANAGED_DISK_STORAGE_ACCOUNT_TYPE=Premium_LRS
export TESTKIT_AZURE_VM_SIZE=Standard_B4ms
export TESTKIT_AZURE_IP_COUNT=100

export TESTKIT_ENGINE=ee-test-19.03

export TESTKIT_AZURE_MACHINE_PASSWORD=****
export TESTKIT_DRIVER=azure

export AZURE_CLIENT_ID=****
export AZURE_CLIENT_SECRET=****
export AZURE_SUBSCRIPTION_ID=****
export AZURE_TENANT_ID=****
```

Go to `orca` repo and run the integration tests with the following command:

```bash
REGISTRY_USERNAME=**** REGISTRY_PASSWORD=**** TESTKIT_PROJECT_TAG=**** TESTKIT_PLATFORM_LINUX=rhel_7.5 \
TESTKIT_PLATFORM_WINDOWS='win_MicrosoftWindowsServer:WindowsServer:2019-Datacenter-Core-with-Containers-smalldisk:latest' \
TESTKIT_PLATFORM_LINUX='rhel_RedHat:RHEL:7.5:latest' TESTKIT_PRESERVE_TEST_MACHINE=1 ORG=dockereng TAG=3.3.0-45d6c4c PULL_IMAGES=1 DEBUG_MODE=1 \
make integration TEST_FLAGS=-v INTEGRATION_TEST_SCOPE=./integration/basic/c1_ww1/...
```

After about 30 minutes, the cluster should be ready to use.

## Demo
Log in to UCP with the creditials `admin`/`or*****`, we should be able to see there are 1 windows node and 1 linux node, both of them are in `ready` state.
Deploy the following service:

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: win-webserver
  name: win-webserver
spec:
  replicas: 2
  template:
    metadata:
      labels:
        app: win-webserver
      name: win-webserver
    spec:
      tolerations:
      - key: "node.kubernetes.io/os"
        operator: "Equal"
        value: "win1809"
        effect: "NoSchedule"
      containers:
      - name: windowswebserver
        image: mcr.microsoft.com/windows/servercore:ltsc2019
        command:
        - powershell.exe
        - -command
        - "<#code used from https://gist.github.com/wagnerandrade/5424431#> ; $$listener = New-Object System.Net.HttpListener ; $$listener.Prefixes.Add('http://*:80/') ; $$listener.Start() ; $$callerCounts = @{} ; Write-Host('Listening at http://*:80/') ; while ($$listener.IsListening) { ;$$context = $$listener.GetContext() ;$$requestUrl = $$context.Request.Url ;$$clientIP = $$context.Request.RemoteEndPoint.Address ;$$response = $$context.Response ;Write-Host '' ;Write-Host('> {0}' -f $$requestUrl) ;  ;$$count = 1 ;$$k=$$callerCounts.Get_Item($$clientIP) ;if ($$k -ne $$null) { $$count += $$k } ;$$callerCounts.Set_Item($$clientIP, $$count) ;$$ip=(Get-NetAdapter | Get-NetIpAddress); $$header='<html><body><H1>Windows Container Web Server</H1>' ;$$callerCountsString='' ;$$callerCounts.Keys | % { $$callerCountsString+='<p>IP {0} callerCount {1} ' -f $$ip[1].IPAddress,$$callerCounts.Item($$_) } ;$$footer='</body></html>' ;$$content='{0}{1}{2}' -f $$header,$$callerCountsString,$$footer ;Write-Output $$content ;$$buffer = [System.Text.Encoding]::UTF8.GetBytes($$content) ;$$response.ContentLength64 = $$buffer.Length ;$$response.OutputStream.Write($$buffer, 0, $$buffer.Length) ;$$response.Close() ;$$responseStatus = $$response.StatusCode ;Write-Host('< {0}' -f $$responseStatus)  } ; "
      nodeSelector:
        beta.kubernetes.io/os: windows
```

Grab an UCP bundle and run the following: 

```
kubectl get nodes
kubectl get pods
kubectl describe pods
```

Now we can try to ping the Windows pods with each other, assuming their IP addresses are `20.0.0.5` and `20.0.0.6`.

```
docker ps | grep servercore
```

and 

```
docker exec -it {id} cmd
```

Now the pods should be able to ping each other by the following Powershell commands:


```powershell
Invoke-WebRequest -Uri "http://20.0.0.5" -UseBasicParsing
Invoke-WebRequest -Uri "http://20.0.0.6" -UseBasicParsing

```

## TODO
Ping between Windows pod and Linux pod

