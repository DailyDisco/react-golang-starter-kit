import { Demo } from "../components/demo/demo";

export function meta() {
    return [
        { title: "React + Go Starter Kit" },
        { name: "description", content: "Welcome to your React + Go full-stack application" },
    ];
}

export default function Home() {
    return <Demo />;
}
