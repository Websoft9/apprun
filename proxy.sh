# For git
export GIT_SSH_COMMAND="ssh -o ProxyCommand='nc -x 127.0.0.1:1088 %h %p'"
export https_proxy=socks5h://127.0.0.1:1088
export http_proxy=socks5h://127.0.0.1:1088