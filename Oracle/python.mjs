import { spawn } from 'child_process';
const a = 'The sun set over the horizon, casting a warm golden glow across the sky. Birds flew in formation, their silhouettes against the fading light. The world seemed to pause for a moment, as if holding its breath before the night arrived'
export async  function GetName(context) {
    return new Promise((resolve, reject) => {
        const pythonProcess = spawn('C:/Users/DonQuixote/.conda/envs/py308/python.exe', ['E:/Code/MLCode/kg_one2set-master/predict.py', context]);

        pythonProcess.stdout.on('data', (data) => {
            if (data.toString().includes("ThisIsName")) {
                const output = data.toString();
                const regex = /ThisIsName\s*(.*)/; 
                const match = output.match(regex);
                if (match) {
                    const partAfterName = match[1].trim();
                    const wordsArray = partAfterName.split(';').map(word => word.trim());
                    resolve(wordsArray[0]);
                }
            }
        });

        pythonProcess.stderr.on('data', (data) => {
            console.error(`Python error: ${data}`);
        });

        pythonProcess.on('close', (code) => {
            console.log(`Python process exited with code ${code}`);
            if (code !== 0) {
                reject(new Error(`Python process failed with code ${code}`));
            }
        });
    });
}
//console.log(await GetName("I love a deer, it's meat is delicious"))
