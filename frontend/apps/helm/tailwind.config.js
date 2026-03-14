/** @type {import('tailwindcss').Config} */
export default {
    content: ["./index.html", "./src/**/*.{vue,ts,tsx}"],
    theme: {
        extend: {
            colors: {
                steel: {
                    50: "#f4f7fb",
                    100: "#e7eef6",
                    200: "#cfdeee",
                    300: "#a8c2df",
                    400: "#7aa0ca",
                    500: "#5d86b6",
                    600: "#4a6c98",
                    700: "#3e587c",
                    800: "#374c67",
                    900: "#314157",
                },
            },
            boxShadow: {
                panel: "0 1px 2px rgba(16, 24, 40, 0.06), 0 1px 3px rgba(16, 24, 40, 0.1)",
            },
        },
    },
    plugins: [],
};
