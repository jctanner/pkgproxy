# DNF
```dnf.conf
[main]
gpgcheck=0
installonly_limit=3
clean_requirements_on_remove=True
best=True
skip_if_unavailable=False

sslverify=false
proxy=http://gosquid:80
```

`dnf --config=/path/to/your/dnf.conf install -y ...`

# PIP
`pip install ... --trusted-host files.pythonhosted.org --trusted-host pypi.org --proxy https://{PROXY_HOST}:{PROXY_PORT}`
