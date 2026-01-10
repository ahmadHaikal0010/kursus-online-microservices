const express = require('express');
const mongoose = require('mongoose');

const app = express();
app.use(express.json());

mongoose.connect(process.env.MONGO_URI)
    .then(() => console.log('MongoDB connected'))
    .catch(err => console.error(err));

app.get('/health', (req, res) => {
    res.json({ status: 'Course Service is Running'});
});

app.listen(3000, () => {
    console.log('Course Service listening on port 3000');
})