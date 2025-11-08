import React, { useEffect, useState } from 'react'



// PlantDetails component
// Props: plantID



function formatISODate(date) {
    // Returns YYYY-MM-DD
    return date.toISOString().slice(0, 10)
}

export default function PlantDetails({ plantID }) {
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


    if (loading) {
        return <div>Loading...</div>
    }

    if (error) {
        return <div className="error">Error: {error}</div>
    }

    return <div>default state now.</div>
}
