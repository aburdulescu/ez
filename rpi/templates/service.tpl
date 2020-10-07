[Unit]
Description={{.description}}
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=1
User=pi
ExecStart={{.execStart}}

[Install]
WantedBy=multi-user.target
