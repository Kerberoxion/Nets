package ru.nsu.fit.krizko.client;

import java.io.*;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.nio.ByteBuffer;

public class Client implements Runnable {

    Socket socket;
    BufferedOutputStream out;
    File file;
    int BUF_SIZE = 4096;

    public Client(InetAddress inetAddress, int port, File file) throws IOException {
        socket = new Socket();
        socket.connect(new InetSocketAddress(inetAddress, port), 10000);
        out = new BufferedOutputStream(socket.getOutputStream());
        this.file = file;
    }

    @Override
    public void run() {
        try {
            sendHeader();
            sendFile();
        }
        catch (IOException ex) {
            System.out.println("Connection to server lost. File didn't send");
        }
        finally {
            try {
                socket.close();
                System.out.println("Socket closed");
            }
            catch (IOException ex) {
                ex.printStackTrace();
            }
        }
    }

    void sendHeader() throws IOException {
        byte[] fileNameBuf = file.getName().getBytes();
        byte[] fileNameLengthBuf = ByteBuffer.allocate(2)
                .putShort((short) fileNameBuf.length).array();
        out.write(fileNameLengthBuf);


        assert (fileNameBuf.length == file.getName().length());
        out.write(fileNameBuf);

        byte[] fileSizeBuf = ByteBuffer.allocate(8).putLong(file.length()).array();
        out.write(fileSizeBuf);

        System.out.println("Header sent");
    }

    void sendFile() throws IOException {
        byte[] fileBuf = new byte[BUF_SIZE];
        long bytesSent = 0;
        long bytesRemain = file.length();
        InputStream inputStream = new FileInputStream(file);
        while (bytesSent < file.length()) {
            int bytesToSend = bytesRemain < BUF_SIZE ? (int) bytesRemain : 4096;
            int bytesReadNow = inputStream.read(fileBuf, 0, bytesToSend);
            out.write(fileBuf, 0, bytesReadNow);
            out.flush();
            bytesSent += bytesReadNow;
            bytesRemain -= bytesReadNow;
            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }
        }
        inputStream.close();
        System.out.println("File sent");
    }
}