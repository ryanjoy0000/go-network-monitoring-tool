
# 🌐 Network Monitoring Tool with Go

Hi there! "Network Monitoring Tool" is a lightweight application developed in Go for real-time monitoring of network performance and health. It provides essential metrics such as latency, bandwidth, and packet loss to ensure optimal network functionality. The tool is designed for easy integration and usability, making it ideal for network administrators and developers.


## 📋 Table of Contents
- [🌟 Purpose](#purpose)
- [📂 Project Structure](#structure)
- [🛠 Installation](#installation) 
- [🤝 Contributing](#contributing)
- [📝 License](#license)
- [🙋 Author](#author)

## <a name="purpose"></a> 🌟 Purpose
1. Monitor Network Health:
- Performance Metrics: Measure things like network speed (latency, bandwidth) and errors (packet loss, connectivity issues).
- Health Status: Determine if the network is up or down and if devices are functioning properly.

## <a name="structure"></a> 📂 Project Structure
We follow a microservice architecture here
1. Backend Service:

    - Data Collectors:
        - Small programs written in Go that run on network devices to gather data.
        - Example: Collecting data on bandwidth usage, latency, and errors.
    - Central Server:
        - A main server that receives data from data collectors.
        - Processes the data

2. Messaging System:
    - Apache Kafka handles the flow of data from the collectors to the central server.

## <a name="installation"></a> 🛠 Installation

### ⚙️ Prerequisites
- **Go**: Ensure you have Go installed on your machine. Verify by running:

```bash
go version
```

If Go is not installed, download it from [the official Go website](https://golang.org/dl/).

- **Docker**: Ensure that Docker is installed and running on your local machine. You can verify the installation by running the following command in your terminal:

  ```bash
  docker --version
  ```

### 📥 Clone the Repository

```bash
git clone https://github.com/ryanjoy0000/go-network-monitoring-tool.git
cd go-network-monitoring-tool
```

### 🏗 Build the Project

```bash
make build
```
## 📝 Future Features
1. Visualize Data with Graphs and Charts: 
    - Use visual tools to make data easy to understand. This helps quickly identify issues and trends.
    - Angular & D3.js to create interactive UI and real-time graphs and charts.

2. Real-time Updates: 
    - Ensure the data displayed is current and updates as new data comes in.
    - Use InfluxDB to store time-series data like network performance metrics over time.
    - Apache Kafka to handle the flow of data from the collectors to the central server.

3. Notification System: 
    - Send alerts via email or SMS when something unusual is detected, like a drop in network speed or an outage.


## <a name="contributing"></a> 🤝 Contributing

I welcome contributions! To get started:

1. 🍴 Fork the repository.
2. 🌿 Create a new branch (`git checkout -b feature/AmazingFeature`).
3. 💾 Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. 📤 Push to the branch (`git push origin feature/AmazingFeature`).
5. 📬 Open a Pull Request.

## <a name="license"></a> 📝 License

This project is licensed under the GNU GENERAL PUBLIC LICENSE


## <a name="author"></a> 🙋Author

**Ryan Joy C.** - [LinkedIn](https://www.linkedin.com/in/ryanjoyc/)
