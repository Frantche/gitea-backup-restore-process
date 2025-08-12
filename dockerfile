FROM registry.access.redhat.com/ubi9

# install mysql
RUN dnf install wget -y \
  && wget https://dev.mysql.com/get/mysql80-community-release-el9-5.noarch.rpm \
  && dnf install ./mysql80-community-release-el9-5.noarch.rpm -y \
  && dnf install mysql -y

# install postgresql
RUN dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-9-x86_64/pgdg-redhat-repo-latest.noarch.rpm \
  && dnf install postgresql -y

COPY gitea_backup_restore gitea_backup_restore
COPY pyproject.toml pyproject.toml
COPY poetry.lock poetry.lock

# install aws CLI
RUN dnf install -y python3 python3-pip \
  && curl -sSL https://install.python-poetry.org | python3 - \
  && /root/.local/bin/poetry install

CMD [ "sleep", "infinity" ]