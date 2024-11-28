package ru.nsu.fit.krizko.server;

import java.io.*;
import java.net.Socket;
import java.nio.ByteBuffer;
import java.util.Timer;

public class ServerThread implements Runnable {

    private final Socket socket;
    private final String fileDir = System.getProperty("user.dir") + "/uploads";
    private String fileName;
    private long fileSize;
    private final InputStream in;
    private long bytesReceived = 0;
    private Timer timer;

    public ServerThread(Socket socket) throws IOException {
        this.socket = socket;
        in = socket.getInputStream();
    }

    @Override
    public void run() {
        try {
            receiveHeader();
            receiveFile();
        }
        catch (IOException | IndexOutOfBoundsException ex) {
            System.out.println("Connection to sender of '" + fileName +
                    "' lost. File didn't receive");
        }
        finally {
            timer.cancel();
            try {
                socket.close();
                System.out.println("Socket of '" + fileName + "' closed");
            }
            catch (IOException ex) {
                ex.printStackTrace();
            }
        }
    }

    void receiveHeader() throws IOException {
        DataInputStream in = new DataInputStream(socket.getInputStream());

        byte[] fileNameLengthBuf = new byte[2];
        in.readFully(fileNameLengthBuf);
        int fileNameLength = ByteBuffer.wrap(fileNameLengthBuf).getShort();

        byte[] fileNameBuf = new byte[fileNameLength];
        in.readFully(fileNameBuf);
        fileName = new String(fileNameBuf);

        byte[] fileSizeBuf = new byte[8];
        in.readFully(fileSizeBuf);
        fileSize = ByteBuffer.wrap(fileSizeBuf).getLong();
        System.out.println("Header received: '" + fileName + "', " + fileSize + " bytes");
    }

    void receiveFile() throws IOException, IndexOutOfBoundsException {
        new File(fileDir).mkdirs();
        File outFile = new File(fileDir + "/" + fileName);

        long startTime;
        try (FileOutputStream outFileStream = new FileOutputStream(outFile)) {

            int BUF_SIZE = 4096;
            byte[] fileBuf = new byte[BUF_SIZE];

            timer = new Timer();
            SpeedCounter speedCounter = new SpeedCounter(this);
            timer.schedule(speedCounter, 1000, 1000);

            long bytesRemain = fileSize;

            startTime = System.currentTimeMillis();
            while (bytesRemain > 0) {
                int bytesReceivedNow = in.read(fileBuf, 0,
                        bytesRemain < BUF_SIZE ? (int) bytesRemain : BUF_SIZE);
                bytesReceived += bytesReceivedNow;
                bytesRemain -= bytesReceivedNow;
                outFileStream.write(fileBuf, 0, bytesReceivedNow);
                outFileStream.flush();

            }
        }catch (IOException err) {
            throw new RuntimeException(err);
        }
        long endTime = System.currentTimeMillis();

        System.out.printf("File '%s' received! Average speed: %.3f Mb/s\n", fileName,
                (fileSize * 1000.0 / (endTime - startTime) / 8 / 1024 / 1024));
    }

    public long getReceivedBytes() {
        return bytesReceived;
    }

    public String getFileName() {
        return fileName;
    }

    public long getFileSize() {
        return fileSize;
    }
}
