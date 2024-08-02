<br />
<div id="readme-top" align="center">
  <a href="https://github.com/itsmrval/ltsninja">
    <img src="https://raw.githubusercontent.com/itsmrval/ltsNinja/main/static/img/logo.svg" alt="Logo" width="120" height="120">
  </a>

  <h3 align="center">ltsNinja</h3>

  <p align="center">
    Simple and lightwell url shortener running GO.
    <br />
    <br />
    <a href="https://lts.ninja">Explore demo</a>
    ·
    <a href="https://github.com/itsmrval/ltsninja/issues">Report Bug</a>
    ·
    <a href="https://github.com/itsmrval/ltsninja/pulls">Pull request</a>
  </p>
</div>


<details>
  <summary>Table of contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">What is ltsNinja ?</a>
      <ul>
        <li><a href="#built-with">Built with</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting started</a>
      <ul>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>



## What is ltsNinja

Homepage             |  Dashboard
:-------------------------:|:-------------------------:
![](https://i.imgur.com/Ykjbxr6.png)  |  ![](https://i.imgur.com/8vhIrax.png)


ltsNinja is a public, self-hosted tool that makes it easy to shorten urls. It's very lightweight and intuitive, which means it runs with very little performance.

Few key points:
* Github login for custom links
* Easy dashboard for users 

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

This section list major frameworks/libraries used

* ![](https://img.shields.io/badge/GO-20232A?style=for-the-badge&logo=go)
* ![](https://img.shields.io/badge/Gin-20232A?style=for-the-badge&logo=gin)
* ![](https://img.shields.io/badge/SqLite-20232A?style=for-the-badge&logo=sqlite&logoColor=blue)
* ![](https://img.shields.io/badge/GitHub%20OAUTH-20232A?style=for-the-badge&logo=github)
* ![](https://img.shields.io/badge/TailWind-20232A?style=for-the-badge&logo=tailwindcss)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



## Getting Started

Now let's see how to set up an ltsNinja instance.

### Installation

1. Create directory
   ```sh
   mkdir /opt/ltsNinja
   cd /opt/ltsNinja
   ```
2. Download the latest release and apply permissions
   ```sh
   wget -O ltsNinja https://github.com/itsmrval/ltsNinja/releases/download/0.1.0/ltsNinja_linux_amd64
   chmod +x ltsNinja
   ```
3. Create the service on systemd
   Write the file
   ```sh
   nano /etc/systemd/system/ltsNinja.service
   ```
   Complete and put the service file below:
   	```txt
   	[Unit]
	Description=LTS Ninja service
	After=network.target
	
	[Service]
	Type=simple
	ExecStart=/opt/ltsNinja/ltsNinja
	Environment="GITHUB_CLIENT_ID=<REPLACE HERE>"
	Environment="GITHUB_CLIENT_SECRET=<REPLACE HERE>"
	Environment="GITHUB_REDIRECT_URL=https://<REPLACE HERE>/callback"
	Environment="DB_PATH=/opt/ltsNinja/database.db"
	Environment="PORT=8080"
	
	[Install]
	WantedBy=multi-user.target
   	```
   
6. Reload systemd and run the service !
   ```sh
   systemctl daemon-reload
   systemctl enable --now ltsNinja
   ```
<p align="right">(<a href="#readme-top">back to top</a>)</p>





## Roadmap

- [x] URL Shortener
- [x] Custom links
- [x] User dashboard
- [x] Edit with github
- [ ] Admin dashboard

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>
