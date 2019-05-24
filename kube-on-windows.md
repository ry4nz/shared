## Introduction

The release of Kubernetes 1.14 includes production support for scheduling Windows containers on Windows nodes in a Kubernetes cluster.  As planned, Docker Enterprise
Universal Control Plane (UCP) will support Kubernetes for Windows containers in Q4 2019.  This document shows how one can configurate and test for this feature in Techinical Preview Release (TPR).

## Minimum Requirement
In the following we show an example for Azure, it will also work on other Cloud platforms such as AWS, or on-prem.  The following table lists the minimum requirement:

| Environment |  Minimum Requirement | 
| :-------------:| :-------------: |
| Memory     |  128GB |
| Size      | Standard_B4ms      |   
| Multiple IPs per VM |  128     |  
| Docker Engine | EE 19.03 |
| Windows Server | 2019 |

## Configuration on Windows Worker Nodes

The following environment variables are set for Windows, they will be changed for the TPR.

```cmd
set TAG=3.3.0-45d6c4c
set ORG=dockereng
```

The following images are required for Windows nodes, one may get them by `docker pull` or an offline bundle.

| Required Windows Images | Remarks |
| :-------------:| -------------  |
| `%ORG%/ucp-kube-binaries-win:%TAG%` | Contains all the binaries required for Kubernetes on Windows |
| `%ORG%/ucp-docker-cni-win:%TAG%` | Contains Docker CNI plugin | 
| `%ORG%/ucp-agent-win:%TAG%` | Also exists for earlier versions of UCP |

```cmd
:: Find the IP address of the windows node
set PRIVATE_IP=
:: Set the username and password for `docker.io`
set REGISTRY_USERNAME=
set REGISTRY_PASSWORD=
```

```cmd
:: Run the following command to add `proxy.local` to `/etc/hosts`:
echo %PRIVATE_IP% proxy.local >> C:\windows\system32\drivers\etc\hosts

:: Copy the required Kubernetes Binaries to desired locations:
docker create --name windowsfiles %ORG%/ucp-kube-binaries-win:%TAG%
docker cp windowsfiles:/bin/containerd.exe .
docker cp windowsfiles:/bin/ctr.exe .
docker cp windowsfiles:/windows/system32/HostNetSvc.dll .
docker cp windowsfiles:/bin/containerd-shim-process-v1.exe .
docker cp windowsfiles:/bin/containerd-shim-process-v1.exe C:\windows\system32\
mkdir C:\k\cni
move HostNetSvc.dll C:\Windows\System32

:: Restart the Windows Server:
Bcdedit /set testsigning on
shutdown /r

:: After the server comes back, start the `containerd` service:
containerd.exe --log-level debug --register-service
powershell -c "Start-Service containerd"

:: Pull the image for containerd
ctr.exe -n com.docker.ucp images pull -u %REGISTRY_USERNAME%:%REGISTRY_PASSWORD% docker.io/%ORG%/ucp-kube-binaries-win:%TAG%
```



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
