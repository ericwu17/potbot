import React, { useEffect, useState } from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'



// PlantDetails component
// Props: plantID, plantName



function formatISODate(date) {
    // Returns YYYY-MM-DD
    return date.toISOString().slice(0, 10)
}

export default function PlantDetails({ plantID, plantName }) {
    const now = new Date()
    const defaultEnd = now.toISOString()
    const defaultStartDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
    const defaultStart = defaultStartDate.toISOString()

    const [startDate, setStartDate] = useState(defaultStart)
    const [endDate, setEndDate] = useState(defaultEnd)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)
    const [logs, setLogs] = useState(null)

    useEffect(() => {
        if (!plantID) {
            setError('Missing plant id')
            return
        }
        setLoading(true)
        setError(null)

        fetch('/api/get_plant_logs', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            body: JSON.stringify({
                plantID,
                startDate: new Date(startDate).toISOString(),
                endDate: new Date(endDate).toISOString(),
            })
        })
            .then(res => {
                if (!res.ok) throw new Error('failed to fetch logs')
                return res.json()
            })
            .then(data => {
                setLogs(data)
            })
            .catch(err => {
                console.warn(err)
                setError('Could not fetch logs')
                setLogs(null)
            })
            .finally(() => setLoading(false))
    }, [plantID, startDate, endDate])


    var details = null;

    if (loading) {
        details = <div>Loading...</div>;
    } else if (error) {
        details = <div className="error">Error: {error}</div>;
    } else {
        console.log(JSON.stringify(logs));
        
        // Transform logs data for Recharts
        const chartData = [];
        const allTimes = new Set();
        
        if (logs) {
            // Collect all unique timestamps
            if (logs.light) logs.light.forEach(entry => allTimes.add(entry.time));
            if (logs.moisture) logs.moisture.forEach(entry => allTimes.add(entry.time));
            if (logs.temp) logs.temp.forEach(entry => allTimes.add(entry.time));
            
            // Create a map for each sensor type by timestamp
            const lightMap = {};
            const moistureMap = {};
            const tempMap = {};
            
            if (logs.light) logs.light.forEach(entry => lightMap[entry.time] = entry.val);
            if (logs.moisture) logs.moisture.forEach(entry => moistureMap[entry.time] = entry.val);
            if (logs.temp) logs.temp.forEach(entry => tempMap[entry.time] = entry.val);
            
            // Build chart data array
            Array.from(allTimes)
                .sort()
                .forEach(time => {
                    const dataPoint = { time };
                    if (lightMap[time] !== undefined) dataPoint.light = lightMap[time];
                    if (moistureMap[time] !== undefined) dataPoint.moisture = moistureMap[time];
                    if (tempMap[time] !== undefined) dataPoint.temp = tempMap[time];
                    chartData.push(dataPoint);
                });
        }
        
        details = (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '30px' }}>
                {/* Light Chart */}
                <div>
                    <h3>Light Intensity</h3>
                    <ResponsiveContainer width="100%" height={300}>
                        <LineChart data={chartData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="time" />
                            <YAxis label={{ value: 'Light (lux)', angle: -90, position: 'insideLeft' }} />
                            <Tooltip />
                            <Legend />
                            <Line type="monotone" dataKey="light" stroke="#8884d8" name="Light Intensity" />
                        </LineChart>
                    </ResponsiveContainer>
                </div>
                
                {/* Moisture Chart */}
                <div>
                    <h3>Soil Moisture</h3>
                    <ResponsiveContainer width="100%" height={300}>
                        <LineChart data={chartData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="time" />
                            <YAxis label={{ value: 'Moisture (%)', angle: -90, position: 'insideLeft' }} />
                            <Tooltip />
                            <Legend />
                            <Line type="monotone" dataKey="moisture" stroke="#82ca9d" name="Soil Moisture" />
                        </LineChart>
                    </ResponsiveContainer>
                </div>
                
                {/* Temperature Chart */}
                <div>
                    <h3>Temperature</h3>
                    <ResponsiveContainer width="100%" height={300}>
                        <LineChart data={chartData}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis dataKey="time" />
                            <YAxis label={{ value: 'Temperature (°C)', angle: -90, position: 'insideLeft' }} />
                            <Tooltip />
                            <Legend />
                            <Line type="monotone" dataKey="temp" stroke="#ffc658" name="Temperature" />
                        </LineChart>
                    </ResponsiveContainer>
                </div>
            </div>
        );
    }



    return (
        <div>
            <div>Data for {plantName}:</div>
            <div>
                TODO: add some toggles and/or inputs here to allow users to customize chart time range...
            </div>
            <div>
                TODO: show some stats about the data (we might want to show the mean or sum?)
            </div>
            {details}
        </div>
    );
}
