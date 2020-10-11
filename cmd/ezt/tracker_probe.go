package main

import "net"

type TrackerProbeClient struct {
	conn *net.UDPConn
}

func NewTrackerProbeClient(address string) (TrackerProbeClient, error) {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return TrackerProbeClient{}, err
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return TrackerProbeClient{}, err
	}
	c := TrackerProbeClient{
		conn: conn,
	}
	return c, nil
}

func (c TrackerProbeClient) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}
