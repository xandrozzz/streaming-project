import {Button, Container, Form} from 'react-bootstrap';
import './App.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import {useState} from "react";
import axios from "axios";

function App() {

    const [newOrder,setNewOrder] = useState("No order could be found.")

    const [findOrder, setFindOrder] = useState("")

    function searchOrder(e) {
        e.preventDefault()

        axios({
            url:'http://localhost:4040/orders/'+findOrder,
            method: "get",
            headers: {
                'Access-Control-Allow-Origin' : '*',
                'Access-Control-Allow-Methods':'GET,PUT,POST,DELETE,PATCH,OPTIONS'
            }
        }).then(function (response) {
            setNewOrder(JSON.stringify(response.data))
        }).catch((e) => {
            console.log(e)
        })
    }

  return (
    <div>
      <Form>
        <Form.Group className="mb-3">
          <Form.Label>Order UID</Form.Label>
          <Form.Control type="text" placeholder="Enter order uid" onChange = {(e) => {setFindOrder(e.target.value)}}/>
          <Form.Text className="text-muted">
            Enter order uid of the order you want to find data for.
          </Form.Text>
        </Form.Group>
        <Button type="submit" onClick={ (e) => {searchOrder(e)} }>Submit</Button>
      </Form>
      <Container>
        <p className="title">Search result:</p>
        <div>{newOrder}</div>
      </Container>
    </div>
  );
}

export default App;
