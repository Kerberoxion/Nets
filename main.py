from time import sleep
import ipaddress
from scapy.all import sniff, send
from scapy.layers.inet import IP, UDP
from scapy.layers.inet6 import IPv6
MULTICAST_PORT = 5005
IPv4addr = True
from threading import Thread
import sys
import socket
def get_local_ip(ip_version=4):
    """Автоматически получает локальный IP-адрес устройства для IPv4 или IPv6."""
    try:
        sock = socket.socket(socket.AF_INET6 if ip_version == 6 else socket.AF_INET, socket.SOCK_DGRAM)
        sock.connect(("8.8.8.8", 80))
        local_ip = sock.getsockname()[0]
        sock.close()
        return local_ip
    except Exception as e:
        print(f"Ошибка при получении локального IP-адреса: {e}")
        return None


def check_multicast_address():
    global IPv4addr
    global dst_ip
    try:
        ip_addr = ipaddress.IPv4Address(dst_ip)
        if(ip_addr.is_multicast):
            return True
        print("адрес не предназначен для мультикаста")
        return False
    except:
        try:
            ip_addr = ipaddress.IPv6Address(dst_ip)
            if (ip_addr.is_multicast):
                dst_ip = dst_ip.replace(dst_ip,ip_addr.exploded)
                IPv4addr = False
                return True
            print("адрес не предназначен для мультикаста")
            return False
        except:
            print("адрес не является валидным")
            return False


def send_multicast_message(interface, dst_ip, IPv4addr, message):
    if(IPv4addr):
        src_ip = "192.168.3.115"
        packet = (
            IP(src=src_ip, dst= dst_ip) /
            UDP(sport=5005, dport=5005) /
            message
        )
    else:
        src_ip = "fe80::682d:2004:a81a:ae76"
        packet = (
            IPv6(src=src_ip, dst= dst_ip) /
            UDP(sport=5005, dport=5005) /
            message
        )
    if src_ip:
        send(packet, iface=interface)
    else:
        print("Не удалось получить исходный IP-адрес.")

def receive_multicast_message(interface, multicast_group):
    """Получает multicast-сообщения с помощью Scapy."""
    print(f"Ожидание сообщений в группе {multicast_group} на интерфейсе {interface}...")
    sniff(iface=interface,prn=packet_callback, store=0)

def packet_callback(packet):
    if packet.haslayer(UDP):
        udp_payload = packet[UDP].payload
        if udp_payload:
            try:
                # Пробуем декодировать как текст
                decoded_message = bytes(udp_payload).decode('utf-8')
                print(f"Получено сообщение: {decoded_message}")
            except UnicodeDecodeError:
                # Если это не текст, выводим байты
                print(f"Получены необработанные данные (байты): {udp_payload}")
        else:
            print("Полезная нагрузка отсутствует в пакете.")
def send_message(interface, dst_ip):
    # Отправляем пакет от имени разных "виртуальных" устройств через разные интерфейсы
    while(True):
        message = "Hi"
        send_multicast_message(interface, dst_ip, IPv4addr, message)
        message = "Hi, world!"
        send_multicast_message(interface, dst_ip, IPv4addr, message)
        message = "Hi,massive world!"
        send_multicast_message(interface, dst_ip, IPv4addr, message)
        sleep(1)

interface = "Ethernet 2"
dst_ip = sys.argv[1]
message = "hello"
if(check_multicast_address()):
    recv_thread = Thread(target=receive_multicast_message, args=(interface, dst_ip))
    send_thread = Thread(target=send_message, args=(interface, dst_ip))
    recv_thread.start()
    send_thread.start()
    recv_thread.join()
    send_thread.join()





