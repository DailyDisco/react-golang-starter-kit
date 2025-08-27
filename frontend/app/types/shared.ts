export interface ExampleData {
  name: string;
  email: string;
  age: number;
}

export interface ApiResponse<T> {
  data: T;
  status: number;
  message: string;
  error?: string;
  success: boolean;
}
