import {Button, ConfigProvider, Input,Layout,Modal,Table,message} from 'antd';
import { Col, Row } from 'antd';
import React, { useEffect, useState } from 'react'
import SearchProduct from '../services/product'
import styles from './product.module.css';
import { Image } from "antd";
import { faSearch } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import "../../app/globals.css";
const ProductView = () => {
    const {Header, Content} = Layout;

    const columns = [
        {
            title: 'Title',
            dataIndex: 'title',
            key: "title"
        },
        {
            title: 'SKU',
            dataIndex: 'sku',
            key: "sku"
        },
        {
            title: 'Description',
            dataIndex: 'description',
            key: "description",
            render: (text,record,index) => {return <Button type='link' onClick={() => {return showModal(record)}}>More Detail</Button>},
        },
    ]

    const [data, setData] = useState({data: null,total: 0});

    const [keyword, setKeyword] = useState();

    const [tableParams, setTableParams] = useState({
        pagination: {
          current: 1,
          pageSize: 10,
          pageSizeOptions:[10,20,50]
        },
      });


    const [isModalOpen, setIsModalOpen] = useState(false);

    const [modalData, setModalData] = useState({title:'',content:''});

    useEffect(()=>{
        if(keyword === undefined || keyword === '' ){
            return;
        }
        fetchData();
    },[keyword, tableParams.pagination.current])


    useEffect(()=>{
        setTableParams({
            ...tableParams,
            pagination: {
                ...tableParams.pagination,
                total: data.total,
            }
        })},[data.total])

    
    function search(value){
        console.log(value.target.value);
        if (keyword==value.target.value){
            fetchData();
        }
        setKeyword(value.target.value);
    }

    async function fetchData(){  
        let currPage = tableParams.pagination.current;
        let currSize = tableParams.pagination.pageSize;
        let from = (currPage-1) * currSize;
        SearchProduct(from, currSize, keyword).then(productData => {setData(productData.data)}).catch(err => {message.error(err.response.data)});
    }

    const handleTableChange = (pagination, filters, sorter) => {
        console.log(tableParams.pagination , pagination);
        if (tableParams.pagination.pageSize != pagination.pageSize){
            pagination.current = 1;
        }
        setTableParams({
          pagination,
          filters,
          ...sorter,
        });
      };

    const showModal = (record) => {
        setModalData({title: record.title, content: record.description});
        setIsModalOpen(true);
    };

    const handleOk = () => {
        setIsModalOpen(false);
    };

    return (
    <ConfigProvider
        theme={{
            components: {
                Input: {
                        activeBg:  "#591C97",
                        hoverBg: "#591C97",
                        fontSize: 32,
                    },
                },
        }}
    >
    <Layout>
            <Row className={styles.header1} align="middle">
                <Col span={4} offset={6}></Col>
                <Col span={4} align="middle"><Image width={"134px"} height={"65px"} preview={false} src="/logo-wb.png"/></Col>
                <Col span={4}></Col>
            </Row>
            <Row className={styles.header2}>
                <Col span={6} offset={6}></Col>
                <Col span={8} align="right"><Input style={{ width: '65%' }} autoComplete='off' maxLength={128} prefix={<FontAwesomeIcon icon={faSearch}></FontAwesomeIcon>} onPressEnter={search} className={styles.search}></Input></Col>
            </Row>
        {/* <Header className={styles.header1}></Header> */}
        {/* <Header className={styles.header2}></Header> */}
        <Content>
            <Table 
                rowKey={record => {return record.id}} 
                columns={columns} 
                dataSource={data.data}
                pagination={tableParams.pagination}
                onChange={handleTableChange}>
            </Table>
            <Modal title={modalData.title} open={isModalOpen} closable={false} width={720} footer={[<Button key="submit" type="primary" onClick={handleOk}>Ok</Button>]}>
                <div dangerouslySetInnerHTML={{__html: `${modalData.content.replaceAll('\n', '</br>')}`}}></div>
            </Modal>
        </Content>
    </Layout>
    </ConfigProvider>
    )
}

export default ProductView;
