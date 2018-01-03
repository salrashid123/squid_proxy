#!/usr/bin/python
# -*- coding: utf8 -*-

import random
import SocketServer
import os
from pyicap import *


deny_list = []

class ThreadingSimpleServer(SocketServer.ThreadingMixIn, ICAPServer):
    pass

class ICAPHandler(BaseICAPRequestHandler):

    def request_OPTIONS(self):
        self.set_icap_response(200)
        self.set_icap_header('Methods', 'REQMOD')
        self.set_icap_header('Service', 'PyICAP Server 1.0')
        self.send_headers(False)

    def request_REQMOD(self):
        self.set_icap_response(200)

        print("======== filter request =======")
        print(self.enc_req)
        print(self.headers)
                
        if (self.enc_req[0] == 'CONNECT' and self.enc_req[1] in deny_list):
            self.set_enc_status('HTTP/1.1 403 Forbidden')
            self.send_headers(False)
            return
 
        if (self.enc_req[1] in deny_list):
            self.set_enc_status('HTTP/1.1 403 Forbidden')
            self.send_headers(False)
            return

        self.no_adaptation_required()


port = 13440


fname = os.getenv('DENY_FILE', '/apps/pyicap/filter_list.txt')
with open(fname) as f:
  content = f.readlines()
  deny_list = [x.strip() for x in content]
print('Initialize deny_list: ' + str(deny_list))


server = ThreadingSimpleServer(('', port), ICAPHandler) 

try:
    while 1:
        server.handle_request()
except KeyboardInterrupt:
    print "Finished"
