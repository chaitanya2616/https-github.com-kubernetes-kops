# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM alpine:3.16

RUN apk add --no-cache \
  bash \
  git \
  git-fast-import \
  openssh \
  python3 \
  python3-dev \
  curl \
  build-base \
  && python3 -m ensurepip \
  && rm -r /usr/lib/python*/ensurepip \
  && pip3 install --upgrade pip setuptools \
  && rm -r /root/.cache \
  && rm -rf /var/cache/apk/*

COPY requirements.txt /requirements.txt
RUN pip install -U -r /requirements.txt

WORKDIR /docs

EXPOSE 3000

COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]
