/*
 * Copyright (c) 2008-2019, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Arrays;

public class BenchServer {
    public static void main(String[] args) throws IOException {
        if (args.length != 2) {
            System.err.println("Usage: BenchServer bindAddress:port responseSize");
            return;
        }
        String[] addrParts = args[0].split(":");
        String host = addrParts[0];
        InetAddress addr = InetAddress.getByName(host);
        int port = Integer.parseInt(addrParts[1]);
        int responseSize = Integer.parseInt(args[1]);
        try (ServerSocket ss = new ServerSocket(port, 50, addr)) {
            System.out.println("ready for clients");
            for (;;) {
                Socket socket = ss.accept();
                System.out.println("client connected");
                new Thread(() -> {
                    try {
                        byte[] response = new byte[responseSize];
                        Arrays.fill(response, (byte) 1);
                        response[response.length - 1] = 0; // zero-terminator
                        byte[] readBuffer = new byte[2048];
                        InputStream is = socket.getInputStream();
                        OutputStream os = socket.getOutputStream();
                        for (;;) {
                            if (!readUntilZero(readBuffer, is)) {
                                return;
                            }
                            os.write(response);
                        }
                    } catch (IOException e) {
                        e.printStackTrace();
                    }
                }).start();
            }
        }
    }

    private static boolean readUntilZero(byte[] readBuffer, InputStream is) throws IOException {
        for (;;) {
            int read = is.read(readBuffer);
            for (int i = 0; i < read; i++) {
                if (readBuffer[i] == 0) {
                    return true;
                }
            }
            if (read < 0) {
                System.out.println("client disconnected");
                return false;
            }
        }
    }
}
